package v1

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/aidar-khabibullin/MSA_Rocket_Factory/order/internal/api/middleware"
	"github.com/aidar-khabibullin/MSA_Rocket_Factory/order/internal/model"
	"github.com/aidar-khabibullin/MSA_Rocket_Factory/platform/pkg/logger"
)

var (
	// errUnauthenticated - в контексте нет аутентифицированного пользователя
	errUnauthenticated = errors.New("unauthenticated")
	// errInvalidSessionUser - UUID пользователя из сессии не удалось разобрать
	errInvalidSessionUser = errors.New("invalid session user")
)

// sessionUserUUID возвращает UUID пользователя, аутентифицированного middleware по сессии
func sessionUserUUID(ctx context.Context) (uuid.UUID, error) {
	sessionUser, ok := middleware.GetUserFromContext(ctx)
	if !ok || sessionUser == nil {
		return uuid.Nil, errUnauthenticated
	}

	userUUID, err := uuid.Parse(sessionUser.GetUuid())
	if err != nil {
		logger.Error(ctx, "❌ некорректный UUID пользователя из сессии",
			zap.String("user_uuid", sessionUser.GetUuid()), zap.Error(err))
		return uuid.Nil, errInvalidSessionUser
	}

	return userUUID, nil
}

// getOwnedOrder загружает заказ и проверяет, что он принадлежит пользователю из сессии.
// Чужой заказ неотличим от несуществующего (model.ErrOrderNotFound): так не раскрывается
// сам факт существования заказа с таким UUID.
func (a *api) getOwnedOrder(ctx context.Context, orderUUID uuid.UUID) (*model.Order, error) {
	userUUID, err := sessionUserUUID(ctx)
	if err != nil {
		return nil, err
	}

	order, err := a.orderService.GetOrder(ctx, orderUUID)
	if err != nil {
		return nil, err
	}

	if order.UserUUID != userUUID {
		logger.Warn(ctx, "🚫 попытка доступа к чужому заказу",
			zap.String("order_uuid", orderUUID.String()),
			zap.String("session_user_uuid", userUUID.String()))
		return nil, model.ErrOrderNotFound
	}

	return order, nil
}
