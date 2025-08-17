package notification

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/aidar-khabibullin/MSA_Rocket_Factory/notification/internal/model"
	"github.com/aidar-khabibullin/MSA_Rocket_Factory/platform/pkg/logger"
)

type service struct {
	telegramClient TelegramClient
}

type TelegramClient interface {
	SendMessage(ctx context.Context, message string) error
}

func NewService(telegramClient TelegramClient) *service {
	return &service{
		telegramClient: telegramClient,
	}
}

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

	err := s.telegramClient.SendMessage(ctx, message)
	if err != nil {
		logger.Error(ctx, "Failed to send OrderPaid notification", zap.Error(err))
		return fmt.Errorf("failed to send notification: %w", err)
	}

	logger.Info(ctx, "OrderPaid notification sent successfully",
		zap.String("order_uuid", event.OrderUUID.String()))

	return nil
}

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

	err := s.telegramClient.SendMessage(ctx, message)
	if err != nil {
		logger.Error(ctx, "Failed to send ShipAssembled notification", zap.Error(err))
		return fmt.Errorf("failed to send notification: %w", err)
	}

	logger.Info(ctx, "ShipAssembled notification sent successfully",
		zap.String("order_uuid", event.OrderUUID.String()))

	return nil
}
