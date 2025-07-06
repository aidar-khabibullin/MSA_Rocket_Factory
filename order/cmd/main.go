package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"

	customMiddleware "github.com/aidar-khabibullin/MSA_Rocket_Factory/order/internal/middleware"
	orderV1 "github.com/aidar-khabibullin/MSA_Rocket_Factory/shared/pkg/openapi/order/v1"
)

const (
	httpPort = "8080"
	// Таймауты для HTTP-сервера
	readHeaderTimeout = 5 * time.Second
	shutdownTimeout   = 10 * time.Second
)

func main() {
	// Создаем хранилище для данных о погоде
	storage := NewOrderStorage()

	// Создаем обработчик API погоды
	orderHandler := NewOrderHandler(storage)

	// Создаем OpenAPI сервер
	orderServer, err := orderV1.NewServer(orderHandler)
	if err != nil {
		log.Fatalf("ошибка создания сервера OpenAPI: %v", err)
	}

	// Инициализируем роутер Chi
	r := chi.NewRouter()

	// Добавляем middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(10 * time.Second))
	r.Use(customMiddleware.RequestLogger)

	// Монтируем обработчики OpenAPI
	r.Mount("/", orderServer)

	// Запускаем HTTP-сервер
	server := &http.Server{
		Addr:              net.JoinHostPort("localhost", httpPort),
		Handler:           r,
		ReadHeaderTimeout: readHeaderTimeout, // Защита от Slowloris атак - тип DDoS-атаки, при которой
		// атакующий умышленно медленно отправляет HTTP-заголовки, удерживая соединения открытыми и истощая
		// пул доступных соединений на сервере. ReadHeaderTimeout принудительно закрывает соединение,
		// если клиент не успел отправить все заголовки за отведенное время.
	}

	// Запускаем сервер в отдельной горутине
	go func() {
		log.Printf("🚀 HTTP-сервер запущен на порту %s\n", httpPort)
		err = server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("❌ Ошибка запуска сервера: %v\n", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Завершение работы сервера...")

	// Создаем контекст с таймаутом для остановки сервера
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	err = server.Shutdown(ctx)
	if err != nil {
		log.Printf("❌ Ошибка при остановке сервера: %v\n", err)
	}

	log.Println("✅ Сервер остановлен")
}

// OrderStorage представляет потокобезопасное хранилище данных для заказов
type OrderStorage struct {
	mu     sync.RWMutex
	orders map[uuid.UUID]*orderV1.OrderDto
}

// NewOrderStorage создает новое хранилище данных для заказов
func NewOrderStorage() *OrderStorage {
	return &OrderStorage{
		orders: make(map[uuid.UUID]*orderV1.OrderDto),
	}
}

// OrderHandler реализует интерфейс orderV1.Handler для обработки запросов к API заказа
type OrderHandler struct {
	storage *OrderStorage
}

// NewOrderHandler создает новый обработчик запросов к API заказа
func NewOrderHandler(storage *OrderStorage) *OrderHandler {
	return &OrderHandler{
		storage: storage,
	}
}

// CreateOrder обрабатывает запрос создания заказа
func (h *OrderHandler) CreateOrder(ctx context.Context, req *orderV1.CreateOrderRequest) (orderV1.CreateOrderRes, error) {
	order := h.storage.CreateOrder(uuid.UUID(req.UserUUID), req.PartUuids)
	if order == nil {
		return &orderV1.InternalServerError{
			Code:    http.StatusInternalServerError,
			Message: "Внутренняя ошибка сервиса",
		}, nil
	}
	return &orderV1.CreateOrderResponse{
		OrderUUID:  orderV1.OrderUUID(order.GetOrderUUID()),
		TotalPrice: orderV1.OptTotalPrice{
			Value: orderV1.TotalPrice(order.GetTotalPrice().Value),
			Set:   order.GetTotalPrice().Set,
		},
	}, nil
}

// GetOrderByUUID обрабатывает запрос получения информации о заказе по UUID
func (h *OrderHandler) GetOrderByUUID(ctx context.Context, params orderV1.GetOrderByUUIDParams) (orderV1.GetOrderByUUIDRes, error) {
	order := h.storage.GetOrder(params.OrderUUID)
	if order == nil {
		return &orderV1.NotFoundError{
			Code:    http.StatusNotFound,
			Message: fmt.Sprint("Не удалось найти заказ с таким UUID: ", params.OrderUUID),
		}, nil
	}
	return order, nil
}

// PostOrderPayment обрабатывает запрос оплаты заказа
func (h *OrderHandler) PayOrder(ctx context.Context, req *orderV1.PayOrderRequest, params orderV1.PayOrderParams) (orderV1.PayOrderRes, error) {
	payOrder := h.storage.PayOrder(params.OrderUUID, int(req.PaymentMethod))
	if payOrder == nil {
		return &orderV1.NotFoundError{
			Code:    http.StatusNotFound,
			Message: fmt.Sprint("Не удалось найти заказ с таким UUID: ", params.OrderUUID),
		}, nil
	}
	return &orderV1.PayOrderResponse{
		TransactionUUID: orderV1.TransactionUUID(payOrder.GetTransactionUUID().Value),
	}, nil
}

// PostOrderCancel обрабатывает запрос отмены заказа
func (h *OrderHandler) CancelOrder(ctx context.Context, params orderV1.CancelOrderParams) (orderV1.CancelOrderRes, error) {
	cancelOrder := h.storage.CancelOrder(params.OrderUUID)
	if cancelOrder == nil {
		return &orderV1.NotFoundError{
			Code:    http.StatusNotFound,
			Message: fmt.Sprint("Не удалось найти заказ с таким UUID: ", params.OrderUUID),
		}, nil
	}
	return nil, nil
}

// NewError создает новую ошибку в формате GenericError
func (h *OrderHandler) NewError(ctx context.Context, err error) *orderV1.GenericErrorStatusCode {
	code := orderV1.OptInt{}
	code.SetTo(http.StatusInternalServerError)

	message := orderV1.OptString{}
	message.SetTo(err.Error())

	return &orderV1.GenericErrorStatusCode{
		StatusCode: http.StatusInternalServerError,
		Response: orderV1.GenericError{
			Code:    code,
			Message: message,
		},
	}
}

// CreateOrder создает заказ
func (s *OrderStorage) CreateOrder(userUUID uuid.UUID, partUUIDs []uuid.UUID) *orderV1.OrderDto {

	newOrder := &orderV1.OrderDto{
		OrderUUID: uuid.New(),
		UserUUID:  userUUID,
		PartUuids: partUUIDs,
		Status:    orderV1.OrderStatus("PENDING_PAYMENT"),
		TotalPrice: orderV1.OptFloat32{
			Value: 0,
			Set:   true,
		},
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.orders[newOrder.OrderUUID] = newOrder
	return newOrder
}

// GetOrder получает информацию о заказе
func (s *OrderStorage) GetOrder(orderUUID uuid.UUID) *orderV1.OrderDto {
	s.mu.RLock()
	defer s.mu.RUnlock()

	order, ok := s.orders[orderUUID]
	if !ok {
		return nil
	}
	return order
}

// PayOrder оплачивает заказ
func (s *OrderStorage) PayOrder(orderUUID uuid.UUID, paymentMethod int) *orderV1.OrderDto {
	payOrder := s.GetOrder(orderUUID)
	if payOrder == nil {
		return nil
	}

	s.mu.Lock()
	payOrder.Status = orderV1.OrderStatus("PAID")
	payOrder.TransactionUUID = orderV1.OptUUID{
		Value: uuid.New(),
		Set:   true,
	}
	payOrder.PaymentMethod = orderV1.OptPaymentMethod{
		Value: orderV1.PaymentMethod(paymentMethod),
		Set:   true,
	}
	defer s.mu.Unlock()

	return payOrder
}

// CancelOrder отменяет заказ
func (s *OrderStorage) CancelOrder(orderUUID uuid.UUID) *orderV1.OrderDto {
	cancelOrder := s.GetOrder(orderUUID)
	if cancelOrder == nil {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	cancelOrder.Status = orderV1.OrderStatus("CANCELLED")

	return nil
}
