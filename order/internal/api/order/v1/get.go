package v1

import (
	"context"
	"errors"
	"net/http"

	"go.uber.org/zap"

	"github.com/aidar-khabibullin/MSA_Rocket_Factory/order/internal/model"
	"github.com/aidar-khabibullin/MSA_Rocket_Factory/platform/pkg/logger"
	orderV1 "github.com/aidar-khabibullin/MSA_Rocket_Factory/shared/pkg/openapi/order/v1"
)

func (a *api) GetOrderByUUID(ctx context.Context, params orderV1.GetOrderByUUIDParams) (orderV1.GetOrderByUUIDRes, error) {
	if a.orderService == nil {
		return &orderV1.InternalServerError{
			Code:    http.StatusInternalServerError,
			Message: "order service not available",
		}, nil
	}

	order, err := a.getOwnedOrder(ctx, params.OrderUUID)
	if err != nil {
		logger.Error(ctx, "Get order error",
			zap.String("order_uuid", params.OrderUUID.String()),
			zap.Error(err),
		)

		if errors.Is(err, errUnauthenticated) {
			return &orderV1.UnauthorizedError{
				Code:    http.StatusUnauthorized,
				Message: "authentication required",
			}, nil
		}

		if errors.Is(err, errInvalidSessionUser) {
			return &orderV1.InternalServerError{
				Code:    http.StatusInternalServerError,
				Message: "invalid session user",
			}, nil
		}

		if errors.Is(err, model.ErrOrderNotFound) {
			return &orderV1.NotFoundError{
				Code:    http.StatusNotFound,
				Message: "order not found",
			}, nil
		}

		if errors.Is(err, model.ErrConvertFromRepo) {
			return &orderV1.InternalServerError{
				Code:    http.StatusInternalServerError,
				Message: "cannot convert order from repository",
			}, nil
		}

		return nil, err
	}

	return &orderV1.OrderDto{
		OrderUUID: order.OrderUUID,
		UserUUID:  order.UserUUID,
		PartUuids: order.PartUUIDs,
		TotalPrice: orderV1.OptFloat32{
			Value: float32(order.TotalPrice),
			Set:   true,
		},
		TransactionUUID: orderV1.OptUUID{
			Value: order.TransactionUUID,
			Set:   true,
		},
		PaymentMethod: orderV1.OptPaymentMethod{
			Value: orderV1.PaymentMethod(order.PaymentMethod),
			Set:   true,
		},
		Status: orderV1.OrderStatus(order.Status),
	}, nil
}
