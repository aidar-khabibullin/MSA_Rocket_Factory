package v1

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/aidar-khabibullin/MSA_Rocket_Factory/order/internal/model"
	"github.com/aidar-khabibullin/MSA_Rocket_Factory/platform/pkg/logger"
	"github.com/aidar-khabibullin/MSA_Rocket_Factory/platform/pkg/tracing"
	orderV1 "github.com/aidar-khabibullin/MSA_Rocket_Factory/shared/pkg/openapi/order/v1"
)

func (a *api) PayOrder(ctx context.Context, req *orderV1.PayOrderRequest, params orderV1.PayOrderParams) (orderV1.PayOrderRes, error) {
	if a.orderService == nil {
		return &orderV1.InternalServerError{
			Code:    http.StatusInternalServerError,
			Message: "order service not available",
		}, nil
	}

	// Создаем root span для операции оплаты заказа
	ctx, span := tracing.StartSpan(ctx, "order.pay_order",
		trace.WithAttributes(
			attribute.String("order_uuid", params.OrderUUID.String()),
			attribute.String("payment_method", string(req.PaymentMethod)),
		),
	)
	defer span.End()

	// Оплатить можно только свой заказ
	if _, err := a.getOwnedOrder(ctx, params.OrderUUID); err != nil {
		return payOrderAccessError(ctx, params.OrderUUID, err)
	}

	orderDraft := model.Order{
		OrderUUID:     params.OrderUUID,
		PaymentMethod: string(req.PaymentMethod),
	}

	order, err := a.orderService.PayOrder(ctx, &orderDraft)
	if err != nil {
		logger.Error(ctx, "Pay order error",
			zap.String("order_uuid", params.OrderUUID.String()),
			zap.String("payment_method", string(req.PaymentMethod)),
			zap.Error(err),
		)

		if errors.Is(err, model.ErrOrderNotFound) {
			return &orderV1.NotFoundError{
				Code:    http.StatusNotFound,
				Message: "order not found",
			}, nil
		}

		if errors.Is(err, model.ErrOrderNotPayable) {
			return &orderV1.ConflictError{
				Code:    http.StatusConflict,
				Message: "order cannot be paid in its current status",
			}, nil
		}

		if errors.Is(err, model.ErrConvertFromRepo) {
			return &orderV1.InternalServerError{
				Code:    http.StatusInternalServerError,
				Message: "cannot convert order from repository",
			}, nil
		}

		if errors.Is(err, model.ErrConvertFromClient) {
			return &orderV1.InternalServerError{
				Code:    http.StatusInternalServerError,
				Message: "can't parse transaction",
			}, nil
		}

		if errors.Is(err, model.ErrPaymentClient) {
			return &orderV1.InternalServerError{
				Code:    http.StatusFailedDependency,
				Message: "payment client error",
			}, nil
		}

		return nil, err
	}

	// Логируем успешную оплату
	logger.Info(ctx, "Order paid successfully",
		zap.String("order_uuid", order.OrderUUID.String()),
		zap.String("payment_method", order.PaymentMethod),
		zap.String("transaction_uuid", order.TransactionUUID.String()),
	)

	return &orderV1.PayOrderResponse{
		TransactionUUID: orderV1.TransactionUUID(order.TransactionUUID),
	}, nil
}

// payOrderAccessError превращает ошибку проверки доступа в ответ API
func payOrderAccessError(ctx context.Context, orderUUID uuid.UUID, err error) (orderV1.PayOrderRes, error) {
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
		logger.Error(ctx, "Pay order: access check error",
			zap.String("order_uuid", orderUUID.String()), zap.Error(err))
		return nil, err
	}
}
