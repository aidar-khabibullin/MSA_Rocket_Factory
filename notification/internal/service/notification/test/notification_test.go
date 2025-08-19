package notification_test

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/aidar-khabibullin/MSA_Rocket_Factory/notification/internal/model"
)

func (s *NotificationServiceSuite) TestNotifyOrderPaid_Success() {
	// Подготавливаем тестовые данные
	event := &model.OrderPaidEvent{
		EventUUID:       uuid.New(),
		OrderUUID:       uuid.New(),
		UserUUID:        uuid.New(),
		PaymentMethod:   "card",
		TransactionUUID: uuid.New(),
	}

	// Настраиваем мок для успешной отправки
	s.telegramClient.On("SendMessage", mock.Anything, int64(0), mock.AnythingOfType("string")).Return(nil)

	// Выполняем тест
	err := s.service.NotifyOrderPaid(context.Background(), event)

	// Проверяем результат
	s.NoError(err)
	s.telegramClient.AssertExpectations(s.T())
}

func (s *NotificationServiceSuite) TestNotifyOrderPaid_TelegramError() {
	// Подготавливаем тестовые данные
	event := &model.OrderPaidEvent{
		EventUUID:       uuid.New(),
		OrderUUID:       uuid.New(),
		UserUUID:        uuid.New(),
		PaymentMethod:   "card",
		TransactionUUID: uuid.New(),
	}

	// Настраиваем мок для ошибки отправки
	expectedError := errors.New("telegram error")
	s.telegramClient.On("SendMessage", mock.Anything, int64(0), mock.AnythingOfType("string")).Return(expectedError)

	// Выполняем тест
	err := s.service.NotifyOrderPaid(context.Background(), event)

	// Проверяем результат
	s.Error(err)
	s.Contains(err.Error(), "failed to send notification")
	s.telegramClient.AssertExpectations(s.T())
}

func (s *NotificationServiceSuite) TestNotifyShipAssembled_Success() {
	// Подготавливаем тестовые данные
	event := &model.ShipAssembledEvent{
		EventUUID: uuid.New(),
		OrderUUID: uuid.New(),
		UserUUID:  uuid.New(),
		BuildTime: 120,
	}

	// Настраиваем мок для успешной отправки
	s.telegramClient.On("SendMessage", mock.Anything, int64(0), mock.AnythingOfType("string")).Return(nil)

	// Выполняем тест
	err := s.service.NotifyShipAssembled(context.Background(), event)

	// Проверяем результат
	s.NoError(err)
	s.telegramClient.AssertExpectations(s.T())
}

func (s *NotificationServiceSuite) TestNotifyShipAssembled_TelegramError() {
	// Подготавливаем тестовые данные
	event := &model.ShipAssembledEvent{
		EventUUID: uuid.New(),
		OrderUUID: uuid.New(),
		UserUUID:  uuid.New(),
		BuildTime: 120,
	}

	// Настраиваем мок для ошибки отправки
	expectedError := errors.New("telegram error")
	s.telegramClient.On("SendMessage", mock.Anything, int64(0), mock.AnythingOfType("string")).Return(expectedError)

	// Выполняем тест
	err := s.service.NotifyShipAssembled(context.Background(), event)

	// Проверяем результат
	s.Error(err)
	s.Contains(err.Error(), "failed to send notification")
	s.telegramClient.AssertExpectations(s.T())
}

func (s *NotificationServiceSuite) TestNotifyOrderPaid_MessageFormat() {
	// Подготавливаем тестовые данные
	event := &model.OrderPaidEvent{
		EventUUID:       uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
		OrderUUID:       uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"),
		UserUUID:        uuid.MustParse("550e8400-e29b-41d4-a716-446655440002"),
		PaymentMethod:   "card",
		TransactionUUID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440003"),
	}

	// Настраиваем мок для проверки формата сообщения
	s.telegramClient.On("SendMessage", mock.Anything, int64(0), mock.MatchedBy(func(message string) bool {
		return strings.Contains(message, "🎉 Заказ оплачен!") &&
			strings.Contains(message, "550e8400-e29b-41d4-a716-446655440001") &&
			strings.Contains(message, "550e8400-e29b-41d4-a716-446655440002") &&
			strings.Contains(message, "card") &&
			strings.Contains(message, "550e8400-e29b-41d4-a716-446655440003")
	})).Return(nil)

	// Выполняем тест
	err := s.service.NotifyOrderPaid(context.Background(), event)

	// Проверяем результат
	s.NoError(err)
	s.telegramClient.AssertExpectations(s.T())
}

func (s *NotificationServiceSuite) TestNotifyShipAssembled_MessageFormat() {
	// Подготавливаем тестовые данные
	event := &model.ShipAssembledEvent{
		EventUUID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
		OrderUUID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"),
		UserUUID:  uuid.MustParse("550e8400-e29b-41d4-a716-446655440002"),
		BuildTime: 120,
	}

	// Настраиваем мок для проверки формата сообщения
	s.telegramClient.On("SendMessage", mock.Anything, int64(0), mock.MatchedBy(func(message string) bool {
		return strings.Contains(message, "🚀 Корабль собран!") &&
			strings.Contains(message, "550e8400-e29b-41d4-a716-446655440001") &&
			strings.Contains(message, "550e8400-e29b-41d4-a716-446655440002") &&
			strings.Contains(message, "120")
	})).Return(nil)

	// Выполняем тест
	err := s.service.NotifyShipAssembled(context.Background(), event)

	// Проверяем результат
	s.NoError(err)
	s.telegramClient.AssertExpectations(s.T())
}
