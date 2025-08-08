package v1

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/aidar-khabibullin/MSA_Rocket_Factory/payment/internal/config"
	"github.com/aidar-khabibullin/MSA_Rocket_Factory/platform/pkg/logger"
	paymentV1 "github.com/aidar-khabibullin/MSA_Rocket_Factory/shared/pkg/proto/payment/v1"
)

type Gateway struct {
	mux    *runtime.ServeMux
	server *http.Server
}

func NewGateway() *Gateway {
	return &Gateway{
		mux: runtime.NewServeMux(),
	}
}

func (g *Gateway) RegisterHandlers(ctx context.Context) error {
	// Создаем подключение к gRPC серверу
	conn, err := grpc.NewClient(
		config.AppConfig().GRPC.Address(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return fmt.Errorf("failed to connect to gRPC server: %w", err)
	}

	// Регистрируем payment service handler
	err = paymentV1.RegisterPaymentServiceHandler(ctx, g.mux, conn)
	if err != nil {
		return fmt.Errorf("failed to register payment service handler: %w", err)
	}

	return nil
}

func (g *Gateway) Start(ctx context.Context) error {
	g.server = &http.Server{
		Addr:              config.AppConfig().Http.Address(),
		Handler:           g.mux,
		ReadHeaderTimeout: 30 * time.Second, // Защита от Slowloris атак
	}

	logger.Info(ctx, fmt.Sprintf("🌐 HTTP Gateway server listening on %s", config.AppConfig().Http.Address()))

	err := g.server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("failed to start HTTP server: %w", err)
	}

	return nil
}

func (g *Gateway) Stop(ctx context.Context) error {
	if g.server != nil {
		return g.server.Shutdown(ctx)
	}
	return nil
}
