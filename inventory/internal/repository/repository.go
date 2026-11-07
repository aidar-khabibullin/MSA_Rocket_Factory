package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/aidar-khabibullin/MSA_Rocket_Factory/inventory/internal/model"
)

type InventoryRepository interface {
	GetPart(ctx context.Context, uuid uuid.UUID) (*model.Part, error)
	ListParts(ctx context.Context, filter *model.Filter) (*[]model.Part, error)

	// Reserve списывает остатки деталей под заказ: повторяющийся uuid означает несколько штук.
	// Операция атомарна (если хотя бы одной детали не хватает, не списывается ничего) и идемпотентна по orderUUID.
	// Возвращает model.ErrPartNotFound, model.ErrInsufficientStock или model.ErrEmptyReservation.
	Reserve(ctx context.Context, orderUUID uuid.UUID, partUUIDs []uuid.UUID) error
	// Release возвращает остатки зарезервированных под заказ деталей. Для заказа без резерва ничего не делает.
	Release(ctx context.Context, orderUUID uuid.UUID) error
}
