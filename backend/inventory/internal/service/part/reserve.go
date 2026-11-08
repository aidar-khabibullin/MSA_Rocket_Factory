package part

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

func (s *service) ReserveParts(ctx context.Context, orderUUID uuid.UUID, partUUIDs []uuid.UUID) error {
	if err := s.repo.Reserve(ctx, orderUUID, partUUIDs); err != nil {
		return fmt.Errorf("service: failed to reserve parts: %w", err)
	}

	return nil
}

func (s *service) ReleaseParts(ctx context.Context, orderUUID uuid.UUID) error {
	if err := s.repo.Release(ctx, orderUUID); err != nil {
		return fmt.Errorf("service: failed to release parts: %w", err)
	}

	return nil
}
