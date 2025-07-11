package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/aidar-khabibullin/MSA_Rocket_Factory/payment/internal/model"
)

type PaymentService interface {
	Pay(ctx context.Context, Order model.Order) (uuid.UUID, error)
}
