package notification_test

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/aidar-khabibullin/MSA_Rocket_Factory/notification/internal/model"
	iamV1 "github.com/aidar-khabibullin/MSA_Rocket_Factory/shared/pkg/proto/iam/v1"
)

// createMockUser создает пользователя с telegram методом оповещения для тестов
func (s *NotificationServiceSuite) createMockUser(userUUID string, chatID int64) *iamV1.User {
	return &iamV1.User{
		Uuid: userUUID,
		Info: &iamV1.UserInfo{
			Login: "testuser",
			Email: "test@example.com",
			NotificationMethods: []*iamV1.NotificationMethod{
				{
					ProviderName: "telegram",
					Target:       strconv.FormatInt(chatID, 10),
				},
			},
		},
	}
}

func (s *NotificationServiceSuite) TestNotifyOrderPaid_Success() {
	// Подготавливаем тестовые данные
	event := &model.OrderPaidEvent{
		EventUUID:       uuid.New(),
		OrderUUID:       uuid.New(),
		UserUUID:        uuid.New(),
		PaymentMethod:   "card",
		TransactionUUID: uuid.New(),
	}

	expectedChatID := int64(12345)

	// Мокаем получение пользователя из IAM
	mockUser := s.createMockUser(event.UserUUID.String(), expectedChatID)
	s.iamClient.On("GetUser", mock.Anything, event.UserUUID.String()).Return(mockUser, nil)

	// Настраиваем мок для успешной отправки
	s.telegramClient.On("SendMessage", mock.Anything, expectedChatID, mock.AnythingOfType("string")).Return(nil)

	// Выполняем тест
	err := s.service.NotifyOrderPaid(context.Background(), event)

	// Проверяем результат
	s.NoError(err)
	s.telegramClient.AssertExpectations(s.T())
	s.iamClient.AssertExpectations(s.T())
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

	expectedChatID := int64(12345)

	// Мокаем получение пользователя из IAM
	mockUser := s.createMockUser(event.UserUUID.String(), expectedChatID)
	s.iamClient.On("GetUser", mock.Anything, event.UserUUID.String()).Return(mockUser, nil)

	// Настраиваем мок для ошибки отправки
	expectedError := errors.New("telegram error")
	s.telegramClient.On("SendMessage", mock.Anything, expectedChatID, mock.AnythingOfType("string")).Return(expectedError)

	// Выполняем тест
	err := s.service.NotifyOrderPaid(context.Background(), event)

	// Проверяем результат
	s.Error(err)
	s.Contains(err.Error(), "failed to send notification")
	s.telegramClient.AssertExpectations(s.T())
	s.iamClient.AssertExpectations(s.T())
}

func (s *NotificationServiceSuite) TestNotifyShipAssembled_Success() {
	// Подготавливаем тестовые данные
	event := &model.ShipAssembledEvent{
		EventUUID: uuid.New(),
		OrderUUID: uuid.New(),
		UserUUID:  uuid.New(),
		BuildTime: 120,
	}

	expectedChatID := int64(67890)

	// Мокаем получение пользователя из IAM
	mockUser := s.createMockUser(event.UserUUID.String(), expectedChatID)
	s.iamClient.On("GetUser", mock.Anything, event.UserUUID.String()).Return(mockUser, nil)

	// Настраиваем мок для успешной отправки
	s.telegramClient.On("SendMessage", mock.Anything, expectedChatID, mock.AnythingOfType("string")).Return(nil)

	// Выполняем тест
	err := s.service.NotifyShipAssembled(context.Background(), event)

	// Проверяем результат
	s.NoError(err)
	s.telegramClient.AssertExpectations(s.T())
	s.iamClient.AssertExpectations(s.T())
}

func (s *NotificationServiceSuite) TestNotifyShipAssembled_TelegramError() {
	// Подготавливаем тестовые данные
	event := &model.ShipAssembledEvent{
		EventUUID: uuid.New(),
		OrderUUID: uuid.New(),
		UserUUID:  uuid.New(),
		BuildTime: 120,
	}

	expectedChatID := int64(67890)

	// Мокаем получение пользователя из IAM
	mockUser := s.createMockUser(event.UserUUID.String(), expectedChatID)
	s.iamClient.On("GetUser", mock.Anything, event.UserUUID.String()).Return(mockUser, nil)

	// Настраиваем мок для ошибки отправки
	expectedError := errors.New("telegram error")
	s.telegramClient.On("SendMessage", mock.Anything, expectedChatID, mock.AnythingOfType("string")).Return(expectedError)

	// Выполняем тест
	err := s.service.NotifyShipAssembled(context.Background(), event)

	// Проверяем результат
	s.Error(err)
	s.Contains(err.Error(), "failed to send notification")
	s.telegramClient.AssertExpectations(s.T())
	s.iamClient.AssertExpectations(s.T())
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

	expectedChatID := int64(11111)

	// Мокаем получение пользователя из IAM
	mockUser := s.createMockUser(event.UserUUID.String(), expectedChatID)
	s.iamClient.On("GetUser", mock.Anything, event.UserUUID.String()).Return(mockUser, nil)

	// Настраиваем мок для проверки формата сообщения
	s.telegramClient.On("SendMessage", mock.Anything, expectedChatID, mock.MatchedBy(func(message string) bool {
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
	s.iamClient.AssertExpectations(s.T())
}

func (s *NotificationServiceSuite) TestNotifyShipAssembled_MessageFormat() {
	// Подготавливаем тестовые данные
	event := &model.ShipAssembledEvent{
		EventUUID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
		OrderUUID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"),
		UserUUID:  uuid.MustParse("550e8400-e29b-41d4-a716-446655440002"),
		BuildTime: 120,
	}

	expectedChatID := int64(22222)

	// Мокаем получение пользователя из IAM
	mockUser := s.createMockUser(event.UserUUID.String(), expectedChatID)
	s.iamClient.On("GetUser", mock.Anything, event.UserUUID.String()).Return(mockUser, nil)

	// Настраиваем мок для проверки формата сообщения
	s.telegramClient.On("SendMessage", mock.Anything, expectedChatID, mock.MatchedBy(func(message string) bool {
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
	s.iamClient.AssertExpectations(s.T())
}

// TestNotifyOrderPaid_IAMError тестирует обработку ошибки получения пользователя из IAM
func (s *NotificationServiceSuite) TestNotifyOrderPaid_IAMError() {
	// Подготавливаем тестовые данные
	event := &model.OrderPaidEvent{
		EventUUID:       uuid.New(),
		OrderUUID:       uuid.New(),
		UserUUID:        uuid.New(),
		PaymentMethod:   "card",
		TransactionUUID: uuid.New(),
	}

	// Настраиваем мок для ошибки IAM
	expectedError := errors.New("iam service error")
	s.iamClient.On("GetUser", mock.Anything, event.UserUUID.String()).Return(nil, expectedError)

	// Выполняем тест
	err := s.service.NotifyOrderPaid(context.Background(), event)

	// Проверяем результат
	s.Error(err)
	s.Contains(err.Error(), "failed to get telegram chatID")
	s.iamClient.AssertExpectations(s.T())
}

// TestNotifyOrderPaid_NoTelegramMethod тестирует случай отсутствия telegram метода оповещения
func (s *NotificationServiceSuite) TestNotifyOrderPaid_NoTelegramMethod() {
	// Подготавливаем тестовые данные
	event := &model.OrderPaidEvent{
		EventUUID:       uuid.New(),
		OrderUUID:       uuid.New(),
		UserUUID:        uuid.New(),
		PaymentMethod:   "card",
		TransactionUUID: uuid.New(),
	}

	// Создаем пользователя без telegram метода оповещения
	mockUser := &iamV1.User{
		Uuid: event.UserUUID.String(),
		Info: &iamV1.UserInfo{
			Login: "testuser",
			Email: "test@example.com",
			NotificationMethods: []*iamV1.NotificationMethod{
				{
					ProviderName: "email",
					Target:       "test@example.com",
				},
			},
		},
	}
	s.iamClient.On("GetUser", mock.Anything, event.UserUUID.String()).Return(mockUser, nil)

	// Выполняем тест
	err := s.service.NotifyOrderPaid(context.Background(), event)

	// Проверяем результат
	s.Error(err)
	s.Contains(err.Error(), "telegram notification method not found")
	s.iamClient.AssertExpectations(s.T())
}
