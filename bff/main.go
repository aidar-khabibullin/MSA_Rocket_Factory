// Package main — тонкий HTTP-шлюз (BFF) для браузерного фронтенда.
// Браузер не умеет gRPC, поэтому /api/auth/* и /api/parts проксируются в iam и inventory.
// Заказы фронтенд вызывает напрямую в order (HTTP).
package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	iamV1 "github.com/aidar-khabibullin/MSA_Rocket_Factory/shared/pkg/proto/iam/v1"
	inventoryV1 "github.com/aidar-khabibullin/MSA_Rocket_Factory/shared/pkg/proto/inventory/v1"
)

const (
	sessionHeader      = "X-Session-Uuid"
	sessionMetadataKey = "session-uuid" // как в platform/pkg/middleware/grpc
	rpcTimeout         = 5 * time.Second
)

type server struct {
	iamConn   *grpc.ClientConn
	invConn   *grpc.ClientConn
	orderAddr string

	auth      iamV1.AuthServiceClient
	users     iamV1.UserServiceClient
	inventory inventoryV1.InventoryServiceClient
}

func main() {
	iamConn := dial(env("IAM_GRPC_ADDR", "localhost:50053"))
	invConn := dial(env("INVENTORY_GRPC_ADDR", "localhost:50051"))

	s := &server{
		iamConn:   iamConn,
		invConn:   invConn,
		orderAddr: env("ORDER_HTTP_ADDR", "http://localhost:8080"),
		auth:      iamV1.NewAuthServiceClient(iamConn),
		users:     iamV1.NewUserServiceClient(iamConn),
		inventory: inventoryV1.NewInventoryServiceClient(invConn),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/auth/register", s.register)
	mux.HandleFunc("POST /api/auth/login", s.login)
	mux.HandleFunc("GET /api/auth/me", s.me)
	mux.HandleFunc("GET /api/parts", s.parts)
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	mux.HandleFunc("GET /api/health", s.health)

	addr := env("BFF_HTTP_ADDR", ":8095")
	log.Printf("bff listening on %s", addr)
	srv := &http.Server{Addr: addr, Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	log.Fatal(srv.ListenAndServe())
}

func dial(addr string) *grpc.ClientConn {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("dial %s: %v", addr, err)
	}
	return conn
}

func (s *server) register(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Login    string `json:"login"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if !decode(w, r, &in) {
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), rpcTimeout)
	defer cancel()

	res, err := s.users.Register(ctx, &iamV1.RegisterRequest{Info: &iamV1.UserRegistrationInfo{
		UserInfo: &iamV1.UserInfo{Login: in.Login, Email: in.Email},
		Password: in.Password,
	}})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"user_uuid": res.GetUserUuid()})
}

func (s *server) login(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}
	if !decode(w, r, &in) {
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), rpcTimeout)
	defer cancel()

	res, err := s.auth.Login(withClientIP(ctx, r), &iamV1.LoginRequest{Login: in.Login, Password: in.Password})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"session_uuid": res.GetSessionUuid()})
}

func (s *server) me(w http.ResponseWriter, r *http.Request) {
	session := r.Header.Get(sessionHeader)
	if session == "" {
		writeError(w, http.StatusUnauthorized, "missing session")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), rpcTimeout)
	defer cancel()

	res, err := s.auth.Whoami(ctx, &iamV1.WhoamiRequest{SessionUuid: session})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"user_uuid": res.GetUser().GetUuid(),
		"login":     res.GetUser().GetInfo().GetLogin(),
		"email":     res.GetUser().GetInfo().GetEmail(),
	})
}

type partDTO struct {
	UUID         string   `json:"part_uuid"`
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Price        float64  `json:"price"`
	Stock        int64    `json:"stock_quantity"`
	Category     string   `json:"category"`
	Manufacturer string   `json:"manufacturer"`
	Tags         []string `json:"tags"`
}

func (s *server) parts(w http.ResponseWriter, r *http.Request) {
	session := r.Header.Get(sessionHeader)
	if session == "" {
		writeError(w, http.StatusUnauthorized, "missing session")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), rpcTimeout)
	defer cancel()
	ctx = metadata.AppendToOutgoingContext(ctx, sessionMetadataKey, session)

	res, err := s.inventory.ListParts(ctx, &inventoryV1.ListPartsRequest{Filter: &inventoryV1.PartsFilter{}})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	out := make([]partDTO, 0, len(res.GetParts()))
	for _, p := range res.GetParts() {
		out = append(out, partDTO{
			UUID:         p.GetPartUuid(),
			Name:         p.GetName(),
			Description:  p.GetDescription(),
			Price:        p.GetPrice(),
			Stock:        p.GetStockQuantity(),
			Category:     p.GetCategory().String(),
			Manufacturer: p.GetManufacturer().GetName(),
			Tags:         p.GetTags(),
		})
	}
	writeJSON(w, http.StatusOK, out)
}

// clientIP возвращает адрес, с которого пришёл запрос. Заголовку X-Forwarded-For не доверяем: его подставляет
// любой клиент, и ограничение попыток входа обходилось бы сменой заголовка.
func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// withClientIP передаёт IP клиента в iam: его rate limiter считает попытки входа по x-forwarded-for.
// Без этого все пользователи веб-интерфейса выглядели бы для iam одним клиентом и делили бы один лимит.
func withClientIP(ctx context.Context, r *http.Request) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "x-forwarded-for", clientIP(r))
}

func decode(w http.ResponseWriter, r *http.Request, dst any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return false
	}
	return true
}

func writeGRPCError(w http.ResponseWriter, err error) {
	if errors.Is(err, context.DeadlineExceeded) {
		writeError(w, http.StatusGatewayTimeout, "upstream timeout")
		return
	}
	st, _ := status.FromError(err)
	code := http.StatusInternalServerError
	switch st.Code() {
	case codes.InvalidArgument:
		code = http.StatusBadRequest
	case codes.Unauthenticated:
		code = http.StatusUnauthorized
	case codes.PermissionDenied:
		code = http.StatusForbidden
	case codes.NotFound:
		code = http.StatusNotFound
	case codes.AlreadyExists:
		code = http.StatusConflict
	case codes.Unavailable:
		code = http.StatusBadGateway
	}
	writeError(w, code, st.Message())
}

func writeError(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]string{"message": msg})
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

type serviceStatus struct {
	Name string `json:"name"`
	OK   bool   `json:"ok"`
}

// health параллельно проверяет зависимости и всегда отвечает 200: недоступность
// сервиса — это данные для индикатора на фронтенде, а не ошибка самого BFF.
func (s *server) health(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	checks := []struct {
		name string
		fn   func(context.Context) bool
	}{
		{"IAM", func(ctx context.Context) bool { return grpcServing(ctx, s.iamConn) }},
		{"Каталог", func(ctx context.Context) bool { return grpcServing(ctx, s.invConn) }},
		{"Заказы", s.orderUp},
	}

	out := make([]serviceStatus, len(checks))
	var wg sync.WaitGroup
	for i, c := range checks {
		wg.Add(1)
		go func() {
			defer wg.Done()
			out[i] = serviceStatus{Name: c.name, OK: c.fn(ctx)}
		}()
	}
	wg.Wait()

	writeJSON(w, http.StatusOK, map[string]any{"services": out})
}

// grpcServing: у inventory health-метод закрыт аутентификацией и отвечает Unauthenticated,
// но это значит, что сервис жив. Недоступным считаем сетевую ошибку и любые другие статусы.
func grpcServing(ctx context.Context, conn *grpc.ClientConn) bool {
	res, err := grpc_health_v1.NewHealthClient(conn).Check(ctx, &grpc_health_v1.HealthCheckRequest{})
	if err != nil {
		return status.Code(err) == codes.Unauthenticated
	}
	return res.GetStatus() == grpc_health_v1.HealthCheckResponse_SERVING
}

// orderUp: /health у order закрыт аутентификацией и отвечает 401, но это значит, что сервис жив.
// Недоступным считаем только сетевую ошибку или 5xx.
func (s *server) orderUp(ctx context.Context) bool {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.orderAddr+"/health", nil)
	if err != nil {
		return false
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return false
	}
	defer res.Body.Close()
	return res.StatusCode < http.StatusInternalServerError
}
