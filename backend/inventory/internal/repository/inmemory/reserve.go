package inmemory

import (
	"context"

	"github.com/google/uuid"

	"github.com/aidar-khabibullin/MSA_Rocket_Factory/inventory/internal/model"
)

func (r *repository) Reserve(_ context.Context, orderUUID uuid.UUID, partUUIDs []uuid.UUID) error {
	if len(partUUIDs) == 0 {
		return model.ErrEmptyReservation
	}

	needed := make(map[string]int64, len(partUUIDs))
	for _, id := range partUUIDs {
		needed[id.String()]++
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, reserved := r.reservations[orderUUID.String()]; reserved {
		return nil // идемпотентность: этот заказ уже зарезервирован
	}

	// Сначала проверяем все детали и только потом списываем: резерв либо целиком, либо никак.
	for id, quantity := range needed {
		part, ok := r.data[id]
		if !ok {
			return model.ErrPartNotFound
		}
		if part.StockQuantity < quantity {
			return model.ErrInsufficientStock
		}
	}

	for id, quantity := range needed {
		r.data[id].StockQuantity -= quantity
	}
	r.reservations[orderUUID.String()] = needed

	return nil
}

func (r *repository) Release(_ context.Context, orderUUID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	reserved, ok := r.reservations[orderUUID.String()]
	if !ok {
		return nil
	}

	for id, quantity := range reserved {
		if part, exists := r.data[id]; exists {
			part.StockQuantity += quantity
		}
	}
	delete(r.reservations, orderUUID.String())

	return nil
}
