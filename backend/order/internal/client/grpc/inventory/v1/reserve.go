package v1

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/aidar-khabibullin/MSA_Rocket_Factory/order/internal/model"
	grpcAuth "github.com/aidar-khabibullin/MSA_Rocket_Factory/platform/pkg/middleware/grpc"
	generaredInventoryV1 "github.com/aidar-khabibullin/MSA_Rocket_Factory/shared/pkg/proto/inventory/v1"
)

func (c inventoryClient) ReserveParts(ctx context.Context, orderUUID uuid.UUID, partUUIDs []uuid.UUID) error {
	ctx = grpcAuth.ForwardSessionUUIDToGRPC(ctx)

	ids := make([]string, len(partUUIDs))
	for i, id := range partUUIDs {
		ids[i] = id.String()
	}

	_, err := c.generatedClient.ReserveParts(ctx, &generaredInventoryV1.ReservePartsRequest{
		OrderUuid: orderUUID.String(),
		PartUuids: ids,
	})

	switch status.Code(err) {
	case codes.OK:
		return nil
	case codes.FailedPrecondition:
		return model.ErrPartsOutOfStock
	case codes.NotFound:
		return model.ErrPartsListNotFound
	default:
		return fmt.Errorf("%w: reserve parts: %w", model.ErrInventoryClient, err)
	}
}

func (c inventoryClient) ReleaseParts(ctx context.Context, orderUUID uuid.UUID) error {
	ctx = grpcAuth.ForwardSessionUUIDToGRPC(ctx)

	if _, err := c.generatedClient.ReleaseParts(ctx, &generaredInventoryV1.ReleasePartsRequest{OrderUuid: orderUUID.String()}); err != nil {
		return fmt.Errorf("%w: release parts: %w", model.ErrInventoryClient, err)
	}

	return nil
}
