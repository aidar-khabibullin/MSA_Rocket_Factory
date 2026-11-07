package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/aidar-khabibullin/MSA_Rocket_Factory/order/internal/model"
)

type OrderRepository interface {
	CreateOrder(ctx context.Context, order *model.Order) (*model.Order, error)
	GetOrder(ctx context.Context, id uuid.UUID) (*model.Order, error)
	UpdateOrder(ctx context.Context, order *model.Order) (*model.Order, error)
	// UpdateOrderIfStatus обновляет заказ, только если его текущий статус равен expected.
	// Если статус уже изменён (например, параллельным запросом), возвращает ErrOrderStatusChanged.
	UpdateOrderIfStatus(ctx context.Context, order *model.Order, expected model.OrderStatus) (*model.Order, error)
}
