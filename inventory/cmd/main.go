package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	inventoryV1API "github.com/aidar-khabibullin/MSA_Rocket_Factory/inventory/internal/api/v1"
	inventoryRepository "github.com/aidar-khabibullin/MSA_Rocket_Factory/inventory/internal/repository/mongo"
	inventoryService "github.com/aidar-khabibullin/MSA_Rocket_Factory/inventory/internal/service/part"
	inventoryV1 "github.com/aidar-khabibullin/MSA_Rocket_Factory/shared/pkg/proto/inventory/v1"
)

const grpcPort = 50051

func main() {
	log.Printf("Inventory service starting...")

	ctx := context.Background()

	// Загружаем переменные окружения
	err := godotenv.Load("../.env")
	if err != nil {
		log.Printf("failed to load .env file: %v\n", err)
		return
	}
	dbURI := os.Getenv("MONGO_URI")
	dbName := os.Getenv("MONGO_INITDB_DATABASE")

	// Создаем клиент MongoDB
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(dbURI))
	if err != nil {
		log.Printf("failed to connect to database: %v\n", err)
		return
	}
	defer func() {
		cerr := client.Disconnect(ctx)
		if cerr != nil {
			log.Printf("failed to disconnect: %v\n", cerr)
		}
	}()

	// Получаем базу данных
	db := client.Database(dbName)

	// Занимаем порт для gRPC сервера
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))
	if err != nil {
		log.Printf("failed to listen: %v\n", err)
		return
	}

	s := grpc.NewServer()

	// Регистрируем сервис
	repo := inventoryRepository.NewRepository(db)
	service := inventoryService.NewService(repo)
	api := inventoryV1API.NewAPI(service)

	inventoryV1.RegisterInventoryServiceServer(s, api)

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
