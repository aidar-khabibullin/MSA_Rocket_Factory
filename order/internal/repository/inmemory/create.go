package inmemory

import (
	"context"

	"github.com/google/uuid"

	"github.com/aidar-khabibullin/MSA_Rocket_Factory/order/internal/model"
	"github.com/aidar-khabibullin/MSA_Rocket_Factory/order/internal/repository/converter"
)

func (r *repository) CreateOrder(ctx context.Context, order *model.Order) (*model.Order, error) {
	// UUID может быть задан заранее (под него уже сделан резерв деталей), иначе генерируем новый
	if order.OrderUUID == uuid.Nil {
		order.OrderUUID = uuid.New()
	}

	// Конвертируем в repo модель ПОСЛЕ установки UUID
	repoOrder := converter.ToRepoOrder(order)

	r.mu.Lock()
	r.data[order.OrderUUID.String()] = *repoOrder
	r.mu.Unlock()

	return order, nil
}
