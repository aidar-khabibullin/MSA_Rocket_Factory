package notification

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/aidar-khabibullin/MSA_Rocket_Factory/notification/internal/client/telegram"
	"github.com/aidar-khabibullin/MSA_Rocket_Factory/notification/internal/model"
	"github.com/aidar-khabibullin/MSA_Rocket_Factory/platform/pkg/logger"
)

type service struct {
	telegramClient telegram.TelegramClient
}

// NewService создает новый сервис уведомлений
func NewService(telegramClient telegram.TelegramClient) *service {
	return &service{
		telegramClient: telegramClient,
	}
}

// NotifyOrderPaid отправляет уведомление об оплате заказа
func (s *service) NotifyOrderPaid(ctx context.Context, event *model.OrderPaidEvent) error {
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

	err := s.telegramClient.SendMessage(ctx, 0, message) // 0 - используем дефолтный chatID
	if err != nil {
		logger.Error(ctx, "Failed to send OrderPaid notification", zap.Error(err))
		return fmt.Errorf("failed to send notification: %w", err)
	}

	logger.Info(ctx, "OrderPaid notification sent successfully",
		zap.String("order_uuid", event.OrderUUID.String()),
		zap.String("payment_method", event.PaymentMethod),
		zap.String("transaction_uuid", event.TransactionUUID.String()))

	return nil
}

// NotifyShipAssembled отправляет уведомление о сборке корабля
func (s *service) NotifyShipAssembled(ctx context.Context, event *model.ShipAssembledEvent) error {
	message := fmt.Sprintf(
		"🚀 Корабль собран!\n"+
			"📦 ID заказа: %s\n"+
			"👤 Пользователь: %s\n"+
			"⏱️ Время сборки: %d сек",
		event.OrderUUID.String(),
		event.UserUUID.String(),
		event.BuildTime,
	)

	err := s.telegramClient.SendMessage(ctx, 0, message) // 0 - используем дефолтный chatID
	if err != nil {
		logger.Error(ctx, "Failed to send ShipAssembled notification", zap.Error(err))
		return fmt.Errorf("failed to send notification: %w", err)
	}

	logger.Info(ctx, "ShipAssembled notification sent successfully",
		zap.String("order_uuid", event.OrderUUID.String()))

	return nil
}
