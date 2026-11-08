package inmemory

import (
	"context"

	"github.com/aidar-khabibullin/MSA_Rocket_Factory/order/internal/model"
	"github.com/aidar-khabibullin/MSA_Rocket_Factory/order/internal/repository/converter"
)

func (r *repository) UpdateOrder(ctx context.Context, order *model.Order) (*model.Order, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Проверяем, существует ли заказ
	_, exists := r.data[order.OrderUUID.String()]
	if !exists {
		return nil, model.ErrOrderNotFound
	}

	// Конвертируем и сохраняем
	repoOrder := converter.ToRepoOrder(order)
	r.data[order.OrderUUID.String()] = *repoOrder

	return order, nil
}

// UpdateOrderIfStatus обновляет заказ, только если его текущий статус равен expected.
func (r *repository) UpdateOrderIfStatus(ctx context.Context, order *model.Order, expected model.OrderStatus) (*model.Order, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	current, exists := r.data[order.OrderUUID.String()]
	if !exists {
		return nil, model.ErrOrderNotFound
	}

	if string(current.Status) != string(expected) {
		return nil, model.ErrOrderStatusChanged
	}

	r.data[order.OrderUUID.String()] = *converter.ToRepoOrder(order)

	return order, nil
}
