package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/aidar-khabibullin/MSA_Rocket_Factory/payment/internal/model"
)

type PaymentService interface {
	// Pay идемпотентна по OrderUuid: повторный вызов возвращает ту же транзакцию.
	Pay(ctx context.Context, Order model.Order) (uuid.UUID, error)
	// Refund возвращает платёж и тоже идемпотентна: повтор возвращает тот же идентификатор возврата.
	Refund(ctx context.Context, orderUUID, transactionUUID uuid.UUID) (uuid.UUID, error)
}
