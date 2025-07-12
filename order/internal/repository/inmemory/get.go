package inmemory

import (
	"context"
	"github.com/google/uuid"
	"github.com/aidar-khabibullin/MSA_Rocket_Factory/order/internal/model"
	"github.com/aidar-khabibullin/MSA_Rocket_Factory/order/internal/repository/converter"
)

func (r *repository) GetOrder(ctx context.Context, id uuid.UUID) (*model.Order, error) {
	r.mu.RLock()
	repoOrder := r.data[id.String()]
	r.mu.RUnlock()

	order, err := converter.RepoToModel(&repoOrder)
	if err != nil {
		return nil, err
	}
	return order, nil
}
