package v1

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/aidar-khabibullin/MSA_Rocket_Factory/order/internal/model"
	"github.com/aidar-khabibullin/MSA_Rocket_Factory/platform/pkg/logger"
	orderV1 "github.com/aidar-khabibullin/MSA_Rocket_Factory/shared/pkg/openapi/order/v1"
)

func (a *api) CreateOrder(ctx context.Context, req *orderV1.CreateOrderRequest) (orderV1.CreateOrderRes, error) {
	logger.Info(ctx, "🔧 CreateOrder: проверяем orderService",
		zap.Bool("orderService_is_nil", a.orderService == nil))

	if a.orderService == nil {
		logger.Error(ctx, "❌ CreateOrder: orderService is nil")
		return &orderV1.InternalServerError{
			Code:    http.StatusInternalServerError,
			Message: "order service not available",
		}, nil
	}

	logger.Info(ctx, "✅ CreateOrder: orderService доступен, продолжаем")

	// Владелец заказа определяется только по сессии: user_uuid из тела запроса клиенту доверять нельзя,
	// иначе любой авторизованный пользователь мог бы создавать заказы от имени другого.
	userUUID, err := sessionUserUUID(ctx)
	if err != nil {
		if errors.Is(err, errUnauthenticated) {
			logger.Warn(ctx, "🚫 CreateOrder: в контексте нет аутентифицированного пользователя")
			return &orderV1.UnauthorizedError{
				Code:    http.StatusUnauthorized,
				Message: "authentication required",
			}, nil
		}

		return &orderV1.InternalServerError{
			Code:    http.StatusInternalServerError,
			Message: "invalid session user",
		}, nil
	}

	if requested := uuid.UUID(req.UserUUID); requested != uuid.Nil && requested != userUUID {
		logger.Warn(ctx, "🚫 CreateOrder: попытка создать заказ от имени другого пользователя",
			zap.String("session_user_uuid", userUUID.String()),
			zap.String("requested_user_uuid", requested.String()))
		return &orderV1.ForbiddenError{
			Code:    http.StatusForbidden,
			Message: "cannot create order for another user",
		}, nil
	}

	orderDraft := model.Order{
		UserUUID:  userUUID,
		PartUUIDs: req.PartUuids,
	}

	createOrder, err := a.orderService.CreateOrder(ctx, &orderDraft)
	if err != nil {
		logger.Error(ctx, "Create order error",
			zap.Any("order", req),
			zap.Error(err),
		)

		if errors.Is(err, model.ErrPartsSpecified) {
			return &orderV1.BadRequestError{
				Code:    http.StatusBadRequest,
				Message: "parts not specified",
			}, nil
		}

		if errors.Is(err, model.ErrPartsListNotFound) {
			return &orderV1.NotFoundError{
				Code:    http.StatusNotFound,
				Message: "parts not found",
			}, nil
		}

		if errors.Is(err, model.ErrPartsOutOfStock) {
			return &orderV1.ConflictError{
				Code:    http.StatusConflict,
				Message: "some parts are out of stock",
			}, nil
		}

		if errors.Is(err, model.ErrConvertFromClient) {
			return &orderV1.InternalServerError{
				Code:    http.StatusInternalServerError,
				Message: "can't parse part",
			}, nil
		}

		if errors.Is(err, model.ErrInventoryClient) {
			return &orderV1.InternalServerError{
				Code:    http.StatusFailedDependency,
				Message: "inventory client error",
			}, nil
		}

		return nil, err
	}

	return &orderV1.CreateOrderResponse{
		OrderUUID: orderV1.OrderUUID(createOrder.OrderUUID),
		TotalPrice: orderV1.OptTotalPrice{
			Value: orderV1.TotalPrice(createOrder.TotalPrice),
			Set:   true,
		},
	}, nil
}
