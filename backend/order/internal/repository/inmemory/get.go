package inmemory

import (
	"context"

	"github.com/google/uuid"

	"github.com/aidar-khabibullin/MSA_Rocket_Factory/order/internal/model"
	"github.com/aidar-khabibullin/MSA_Rocket_Factory/order/internal/repository/converter"
)

func (r *repository) GetOrder(ctx context.Context, id uuid.UUID) (*model.Order, error) {
	r.mu.RLock()
	repoOrder, exists := r.data[id.String()]
	r.mu.RUnlock()

	if !exists {
		return nil, model.ErrOrderNotFound
	}

	order, err := converter.ToModelOrder(&repoOrder)
	if err != nil {
		return nil, err
	}

	return order, nil
}
