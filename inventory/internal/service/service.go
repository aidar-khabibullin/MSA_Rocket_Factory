package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/aidar-khabibullin/MSA_Rocket_Factory/inventory/internal/model"
)

type InventoryService interface {
	GetPart(ctx context.Context, uuid uuid.UUID) (*model.Part, error)
	ListParts(ctx context.Context, filter *model.Filter) (*[]model.Part, error)
	ReserveParts(ctx context.Context, orderUUID uuid.UUID, partUUIDs []uuid.UUID) error
	ReleaseParts(ctx context.Context, orderUUID uuid.UUID) error
}
