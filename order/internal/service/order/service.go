package order

import (
	"github.com/aidar-khabibullin/MSA_Rocket_Factory/order/internal/client/grpc"
	"github.com/aidar-khabibullin/MSA_Rocket_Factory/order/internal/repository"
	def "github.com/aidar-khabibullin/MSA_Rocket_Factory/order/internal/service"
)

var _ def.OrderService = (*service)(nil)

type service struct {
	orderRepository repository.OrderRepository
	inventoryClient grpc.InventoryClient
	paymentClient   grpc.PaymentClient
}

func NewService(
	orderRepository repository.OrderRepository,
	inventoryClient grpc.InventoryClient,
	paymentClient grpc.PaymentClient,
) *service {
	return &service{
		orderRepository: orderRepository,
		inventoryClient: inventoryClient,
		paymentClient:   paymentClient,
	}
}
