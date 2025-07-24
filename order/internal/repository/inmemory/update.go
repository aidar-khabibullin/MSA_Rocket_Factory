package inmemory

import (
	"context"

	"github.com/aidar-khabibullin/MSA_Rocket_Factory/order/internal/model"
	"github.com/aidar-khabibullin/MSA_Rocket_Factory/order/internal/repository/converter"
)

func (r *repository) UpdateOrder(ctx context.Context, order *model.Order) (*model.Order, error) {
	repoOrder := converter.ToRepoOrder(order)

	r.mu.Lock()
	r.data[order.OrderUUID.String()] = *repoOrder
	r.mu.Unlock()

	return order, nil
}
