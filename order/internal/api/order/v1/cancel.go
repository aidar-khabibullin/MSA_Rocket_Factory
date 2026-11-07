package v1

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/samber/lo"
	"go.uber.org/zap"

	"github.com/aidar-khabibullin/MSA_Rocket_Factory/order/internal/model"
	"github.com/aidar-khabibullin/MSA_Rocket_Factory/platform/pkg/logger"
	orderV1 "github.com/aidar-khabibullin/MSA_Rocket_Factory/shared/pkg/openapi/order/v1"
)

func (a *api) CancelOrder(ctx context.Context, params orderV1.CancelOrderParams) (orderV1.CancelOrderRes, error) {
	if a.orderService == nil {
		return &orderV1.InternalServerError{
			Code:    http.StatusInternalServerError,
			Message: "order service not available",
		}, nil
	}

	// Отменить можно только свой заказ
	if _, err := a.getOwnedOrder(ctx, params.OrderUUID); err != nil {
		return cancelOrderAccessError(ctx, params.OrderUUID, err)
	}

	orderDraft := model.Order{
		OrderUUID: params.OrderUUID,
	}

	_, err := a.orderService.CancelOrder(ctx, lo.ToPtr(orderDraft))
	if err != nil {
		logger.Error(ctx, "Cancel order error",
			zap.String("order_uuid", params.OrderUUID.String()),
			zap.Error(err),
		)

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

		if errors.Is(err, model.ErrPaid) {
			return &orderV1.BadRequestError{
				Code:    http.StatusBadRequest,
				Message: "order already paid",
			}, nil
		}

		if errors.Is(err, model.ErrCancelled) {
			return &orderV1.BadRequestError{
				Code:    http.StatusBadRequest,
				Message: "order already cancelled",
			}, nil
		}

		return nil, err
	}

	return &orderV1.CancelOrderNoContent{}, nil
}

// cancelOrderAccessError превращает ошибку проверки доступа в ответ API
func cancelOrderAccessError(ctx context.Context, orderUUID uuid.UUID, err error) (orderV1.CancelOrderRes, error) {
	switch {
	case errors.Is(err, errUnauthenticated):
		return &orderV1.UnauthorizedError{Code: http.StatusUnauthorized, Message: "authentication required"}, nil
	case errors.Is(err, errInvalidSessionUser):
		return &orderV1.InternalServerError{Code: http.StatusInternalServerError, Message: "invalid session user"}, nil
	case errors.Is(err, model.ErrOrderNotFound):
		return &orderV1.NotFoundError{Code: http.StatusNotFound, Message: "order not found"}, nil
	case errors.Is(err, model.ErrConvertFromRepo):
		return &orderV1.InternalServerError{Code: http.StatusInternalServerError, Message: "cannot convert order from repository"}, nil
	default:
		logger.Error(ctx, "Cancel order: access check error",
			zap.String("order_uuid", orderUUID.String()), zap.Error(err))
		return nil, err
	}
}
