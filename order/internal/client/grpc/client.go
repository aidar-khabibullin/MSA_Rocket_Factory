package grpc

import (
	"context"

	"github.com/google/uuid"

	"github.com/aidar-khabibullin/MSA_Rocket_Factory/order/internal/model"
)

type InventoryClient interface {
	ListParts(ctx context.Context, filter *model.Filter) (*[]model.Part, error)
	// ReserveParts резервирует детали под заказ (идемпотентно). Возвращает model.ErrPartsOutOfStock, если остатка не хватает.
	ReserveParts(ctx context.Context, orderUUID uuid.UUID, partUUIDs []uuid.UUID) error
	// ReleaseParts снимает резерв заказа и возвращает остатки (идемпотентно).
	ReleaseParts(ctx context.Context, orderUUID uuid.UUID) error
}

type PaymentClient interface {
	// CreatePayment идемпотентен по заказу: повторный вызов возвращает ту же транзакцию.
	CreatePayment(ctx context.Context, order *model.Order) (*model.Order, error)
	// RefundPayment возвращает платёж заказа (компенсация, если заказ не удалось обновить после оплаты).
	RefundPayment(ctx context.Context, order *model.Order) error
}
