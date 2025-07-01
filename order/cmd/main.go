package main

import (
	"context"
	"errors"
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

type Order struct {
	OrderUUID uuid.UUID
	UserUUID uuid.UUID
	PartUUIDs []uuid.UUID
	TotalPrice float64
	TransactionUUID uuid.UUID
	PaymentMethod string
	Status string
}

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
	mu       sync.RWMutex
	orders map[string]*Order
}

// NewOrderStorage создает новое хранилище данных для заказов
func NewOrderStorage() *OrderStorage {
	return &OrderStorage{
		orders: make(map[string]*Order),
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

// PostOrder обрабатывает запрос создания заказа
func (h *OrderHandler) PostOrder(ctx context.Context, req *orderV1.CreateOrderRequest) (orderV1.CreateOrderRes, error) {
	order := h.storage.CreateOrder(req.UserUUID, req.PartUUIDs)
	if order == nil {
		return &orderV1.InternalServerError{
			StatusCode: http.StatusInternalServerError,
			Response:   orderV1.InternalServerErrorResponse{
				Code:    http.StatusInternalServerError,
				Message: "Внутренняя ошибка сервиса",
		},
		}, nil

	}
	return order, nil
}

// GetOrderByUUID обрабатывает запрос получения информации о заказе по UUID
func (h *OrderHandler) GetOrderByUUID(ctx context.Context, params orderV1.GetOrderByUUIDParams) (orderV1.GetOrderByUUIDRes, error) {
	order := h.storage.GetOrder(params.OrderUUID)
	if order == nil {
		return &orderV1.NotFoundError{
			StatusCode: http.StatusNotFound,
			Response:   orderV1.NotFoundErrorResponse{
				Code:    http.StatusNotFound,
				Message: "Не удалось найти заказ с таким UUID: " + params.OrderUUID,
		},
		}, nil

	}
	return order, nil
}

// PostOrderPayment обрабатывает запрос оплаты заказа
func (h *OrderHandler) PostOrderPayment(ctx context.Context, req *orderV1.PayOrderRequest, params orderV1.PayOrderParams) (orderV1.PayOrderRes, error) {
	order := h.storage.PayOrder(params.OrderUUID, req.PaymentMethod)
	if order == nil {
		return &orderV1.NotFoundError{
			StatusCode: http.StatusNotFound,
			Response:   orderV1.NotFoundErrorResponse{
				Code:    http.StatusNotFound,
				Message: "Не удалось найти заказ с таким UUID: " + params.OrderUUID,
		},
		}, nil

	}
	return order, nil
}

// PostOrderCancel обрабатывает запрос отмены заказа
func (h *OrderHandler) PostOrderCancel(ctx context.Context, params orderV1.CancelOrderParams) (orderV1.CancelOrderRes, error) {
	order:= h.storage.CancelOrder(params.OrderUUID)
	if order == nil {

	}
	return order, nil
}

// NewError создает новую ошибку в формате GenericError
func (h *OrderHandler) NewError(ctx context.Context, err error) *orderV1.GenericErrorStatusCode {
	return &orderV1.GenericErrorStatusCode{
		StatusCode: http.StatusInternalServerError,
		Response:   orderV1.GenericErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		},
	}	
}

// CreateOrder создает заказ
func (s *OrderStorage) CreateOrder(userUUID string, partUUIDs []string) *Order {
/*
Поведение:
	- Получает детали через InventoryService.ListParts.
	- Проверяет, что все детали существуют. Если хотя бы одной нет — возвращает ошибку.
	- Считает total_price.
	- Генерирует order_uuid.
	- Сохраняет заказ со статусом PENDING_PAYMENT.
*/
	addOrder := &orderV1.Order{
		OrderUUID: uuid.NewString(),
		UserUUID: userUUID,
		PartUUIDs: partUUIDs,
		TotalPrice: 100,
		Status: orderV1.OrderStatusPendingPayment,
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.orders[uuid.NewString()] = addOrder
	return addOrder
}

// GetOrder получает информацию о заказе
func (s *OrderStorage) GetOrder(orderUUID string) *Order {
	s.mu.RLock()
	defer s.mu.RUnlock()

	order, ok := s.orders[orderUUID]
	if !ok {
		return nil
	}
	return order
}

// PayOrder оплачивает заказ
func (s *OrderStorage) PayOrder(orderUUID string, paymentMethod orderV1.PaymentMethod) *Order {
/*
Поведение:
	- Находит заказ по order_uuid. Если не существует — возвращает 404 Not Found.
	- Вызывает PaymentService.PayOrder, передаёт user_uuid, order_uuid и payment_method. Получаетtransaction_uuid.
	- Обновляет заказ: статус → PAID, сохраняет transaction_uuid, payment_method.
*/
	order:= s.GetOrder(orderUUID)
	if order == nil {
		return nil
	}
	
	return order
}

// CancelOrder отменяет заказ
func (s *OrderStorage) CancelOrder(orderUUID string) *Order {	
/*
Поведение:
	- Проверяет статус заказа.
	- Если PENDING_PAYMENT — меняет статус на CANCELLED.
	- Если PAID — возвращает ошибку 409.
*/
	order:= s.GetOrder(orderUUID)
	if order == nil {
		return nil
	}
	
	return order
}
	
	
	