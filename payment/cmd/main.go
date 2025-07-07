package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	paymentV1 "github.com/aidar-khabibullin/MSA_Rocket_Factory/shared/pkg/proto/payment/v1"
)

const grpcPort = 50052

// PaymentService реализует интерфейс paymentV1.Service для обработки запросов к API платежа
type paymentService struct {
	paymentV1.UnimplementedPaymentServiceServer
}

func main() {
	log.Printf("Payment service starting...")
	// Создаем gRPC соединение
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))
	if err != nil {
		log.Printf("failed to listen: %v\n", err)
		return
	}
	defer func() {
		if cerr := lis.Close(); cerr != nil {
			log.Printf("failed to close listener: %v\n", cerr)
		}
	}()

	// Создаем gRPC сервер
	s := grpc.NewServer()

	// Регистрируем наш сервис
	service := &paymentService{}
	paymentV1.RegisterPaymentServiceServer(s, service)

	// Включаем рефлексию для отладки
	reflection.Register(s)

	go func() {
		log.Printf("🚀 gRPC server listening on %d\n", grpcPort)
		err = s.Serve(lis)
		if err != nil {
			log.Printf("failed to serve: %v\n", err)
			return
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("🛑 Shutting down gRPC server...")
	s.GracefulStop()
	log.Println("✅ Server stopped")
}

// PayOrder оплачивает заказ
func (s *paymentService) PayOrder(ctx context.Context, req *paymentV1.PayOrderRequest) (*paymentV1.PayOrderResponse, error) {
	transaction_uuid := uuid.New()
	log.Printf("Оплата прошла успешно, transaction_uuid: %s\n", transaction_uuid)
	return &paymentV1.PayOrderResponse{
		TransactionUuid: transaction_uuid.String(),
	}, nil
}
