package v1

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/aidar-khabibullin/MSA_Rocket_Factory/inventory/internal/model"
	"github.com/aidar-khabibullin/MSA_Rocket_Factory/platform/pkg/logger"
	inventoryV1 "github.com/aidar-khabibullin/MSA_Rocket_Factory/shared/pkg/proto/inventory/v1"
)

func (a *api) ReserveParts(ctx context.Context, req *inventoryV1.ReservePartsRequest) (*inventoryV1.ReservePartsResponse, error) {
	orderUUID, err := uuid.Parse(req.GetOrderUuid())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid order uuid")
	}

	partUUIDs := make([]uuid.UUID, 0, len(req.GetPartUuids()))
	for _, raw := range req.GetPartUuids() {
		id, err := uuid.Parse(raw)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid part uuid")
		}
		partUUIDs = append(partUUIDs, id)
	}

	err = a.inventoryService.ReserveParts(ctx, orderUUID, partUUIDs)
	switch {
	case err == nil:
		return &inventoryV1.ReservePartsResponse{}, nil
	case errors.Is(err, model.ErrEmptyReservation):
		return nil, status.Error(codes.InvalidArgument, "reservation must contain at least one part")
	case errors.Is(err, model.ErrPartNotFound):
		return nil, status.Error(codes.NotFound, "part not found")
	case errors.Is(err, model.ErrInsufficientStock):
		// FailedPrecondition: повтор запроса не поможет, пока на складе не появятся детали.
		return nil, status.Error(codes.FailedPrecondition, "insufficient stock")
	default:
		logger.Error(ctx, "Failed to reserve parts", zap.String("order_uuid", orderUUID.String()), zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to reserve parts")
	}
}

func (a *api) ReleaseParts(ctx context.Context, req *inventoryV1.ReleasePartsRequest) (*inventoryV1.ReleasePartsResponse, error) {
	orderUUID, err := uuid.Parse(req.GetOrderUuid())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid order uuid")
	}

	if err := a.inventoryService.ReleaseParts(ctx, orderUUID); err != nil {
		logger.Error(ctx, "Failed to release parts", zap.String("order_uuid", orderUUID.String()), zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to release parts")
	}

	return &inventoryV1.ReleasePartsResponse{}, nil
}
