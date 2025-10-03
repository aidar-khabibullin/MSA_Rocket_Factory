package notification

import (
	"context"
	"fmt"
	"strconv"

	"go.uber.org/zap"

	"github.com/aidar-khabibullin/MSA_Rocket_Factory/notification/internal/client/iam"
	"github.com/aidar-khabibullin/MSA_Rocket_Factory/notification/internal/client/telegram"
	"github.com/aidar-khabibullin/MSA_Rocket_Factory/notification/internal/model"
	"github.com/aidar-khabibullin/MSA_Rocket_Factory/platform/pkg/logger"
)

type service struct {
	telegramClient telegram.TelegramClient
	iamClient      iam.Client
}

func NewService(ctx context.Context, telegramClient telegram.TelegramClient, iamClient iam.Client) *service {
	svc := &service{
		telegramClient: telegramClient,
		iamClient:      iamClient,
	}

	logger.Info(ctx, "Сервис уведомлений создан",
		zap.Bool("iam_client_available", iamClient != nil))

	return svc
}

func (s *service) getTelegramChatID(ctx context.Context, userUUID string) (int64, error) {
	user, err := s.iamClient.GetUser(ctx, userUUID)
	if err != nil {
		logger.Error(ctx, "Ошибка получения данных пользователя из IAM",
			zap.Error(err), zap.String("user_uuid", userUUID))
		return 0, fmt.Errorf("failed to get user from IAM: %w", err)
	}

	if user.GetInfo() == nil {
		logger.Error(ctx, "Отсутствует информация о пользователе",
			zap.String("user_uuid", userUUID))
		return 0, fmt.Errorf("user info is nil")
	}

	notificationMethods := user.GetInfo().GetNotificationMethods()
	for _, method := range notificationMethods {
		if method.GetProviderName() == "telegram" {
			chatID, err := strconv.ParseInt(method.GetTarget(), 10, 64)
			if err != nil {
				logger.Error(ctx, "Не удалось преобразовать telegram target в chatID",
					zap.Error(err),
					zap.String("user_uuid", userUUID),
					zap.String("target", method.GetTarget()))
				return 0, fmt.Errorf("failed to parse telegram chatID: %w", err)
			}

			return chatID, nil
		}
	}

	logger.Warn(ctx, "Telegram метод оповещения не найден для пользователя",
		zap.String("user_uuid", userUUID))
	return 0, fmt.Errorf("telegram notification method not found for user %s", userUUID)
}

func (s *service) NotifyOrderPaid(ctx context.Context, event *model.OrderPaidEvent) error {
	chatID, err := s.getTelegramChatID(ctx, event.UserUUID.String())
	if err != nil {
		logger.Error(ctx, "Не удалось получить telegram chatID для пользователя",
			zap.Error(err),
			zap.String("user_uuid", event.UserUUID.String()))
		return fmt.Errorf("failed to get telegram chatID: %w", err)
	}

	message := fmt.Sprintf(
		"🎉 Заказ оплачен!\n"+
			"📦 ID заказа: %s\n"+
			"👤 Пользователь: %s\n"+
			"💳 Способ оплаты: %s\n"+
			"🔑 Транзакция: %s",
		event.OrderUUID.String(),
		event.UserUUID.String(),
		event.PaymentMethod,
		event.TransactionUUID.String(),
	)

	err = s.telegramClient.SendMessage(ctx, chatID, message)
	if err != nil {
		logger.Error(ctx, "Failed to send OrderPaid notification", zap.Error(err))
		return fmt.Errorf("failed to send notification: %w", err)
	}

	logger.Info(ctx, "OrderPaid notification sent successfully",
		zap.String("order_uuid", event.OrderUUID.String()),
		zap.String("user_uuid", event.UserUUID.String()),
		zap.Int64("chat_id", chatID))

	return nil
}

func (s *service) NotifyShipAssembled(ctx context.Context, event *model.ShipAssembledEvent) error {
	chatID, err := s.getTelegramChatID(ctx, event.UserUUID.String())
	if err != nil {
		logger.Error(ctx, "Не удалось получить telegram chatID для пользователя",
			zap.Error(err),
			zap.String("user_uuid", event.UserUUID.String()))
		return fmt.Errorf("failed to get telegram chatID: %w", err)
	}

	message := fmt.Sprintf(
		"🚀 Корабль собран!\n"+
			"📦 ID заказа: %s\n"+
			"👤 Пользователь: %s\n"+
			"⏱️ Время сборки: %d сек",
		event.OrderUUID.String(),
		event.UserUUID.String(),
		event.BuildTime,
	)

	err = s.telegramClient.SendMessage(ctx, chatID, message)
	if err != nil {
		logger.Error(ctx, "Failed to send ShipAssembled notification", zap.Error(err))
		return fmt.Errorf("failed to send notification: %w", err)
	}

	logger.Info(ctx, "ShipAssembled notification sent successfully",
		zap.String("order_uuid", event.OrderUUID.String()),
		zap.String("user_uuid", event.UserUUID.String()),
		zap.Int64("chat_id", chatID))

	return nil
}
