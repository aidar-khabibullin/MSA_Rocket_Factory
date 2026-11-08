package order_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/aidar-khabibullin/MSA_Rocket_Factory/order/internal/model"
)

func (s *ServiceSuite) TestPayOrder_Success() {
	// Тестовые данные
	orderUUID := uuid.New()
	userUUID := uuid.New()
	partUUID := uuid.New()
	transactionUUID := uuid.New()

	// Входящий запрос с методом оплаты
	incomingOrder := &model.Order{
		OrderUUID:     orderUUID,
		PaymentMethod: "CARD",
	}

	// Заказ из БД (без метода оплаты)
	dbOrder := &model.Order{
		OrderUUID:  orderUUID,
		UserUUID:   userUUID,
		PartUUIDs:  []uuid.UUID{partUUID},
		TotalPrice: 100.0,
		Status:     model.StatusPendingPayment,
	}

	// Заказ после оплаты
	paidOrder := &model.Order{
		OrderUUID:       orderUUID,
		UserUUID:        userUUID,
		PartUUIDs:       []uuid.UUID{partUUID},
		TotalPrice:      100.0,
		TransactionUUID: transactionUUID,
		PaymentMethod:   "CARD", // Должен быть установлен из входящего запроса
		Status:          model.StatusPaid,
	}

	// Настройка моков
	s.orderRepository.On("GetOrder", mock.Anything, orderUUID).
		Return(dbOrder, nil)

	// Создаем заказ с установленным PaymentMethod для передачи в CreatePayment
	orderWithPaymentMethod := &model.Order{
		OrderUUID:     orderUUID,
		UserUUID:      userUUID,
		PartUUIDs:     []uuid.UUID{partUUID},
		TotalPrice:    100.0,
		Status:        model.StatusPendingPayment,
		PaymentMethod: "CARD", // Установлен из входящего запроса
	}

	s.paymentClient.On("CreatePayment", mock.Anything, orderWithPaymentMethod).
		Return(paidOrder, nil)
	s.orderRepository.On("UpdateOrderIfStatus", mock.Anything, paidOrder, model.StatusPendingPayment).
		Return(paidOrder, nil)

	// Вызов метода
	result, err := s.service.PayOrder(context.Background(), incomingOrder)

	// Проверка результата
	s.Require().NoError(err)
	s.Require().NotNil(result)
	s.Require().Equal(model.StatusPaid, result.Status)
	s.Require().Equal(transactionUUID, result.TransactionUUID)
	s.Require().Equal("CARD", result.PaymentMethod)

	s.orderRepository.AssertExpectations(s.T())
	s.paymentClient.AssertExpectations(s.T())
}

func (s *ServiceSuite) TestPayOrder_OrderNotFound() {
	// Тестовые данные
	orderUUID := uuid.New()
	incomingOrder := &model.Order{
		OrderUUID:     orderUUID,
		PaymentMethod: "CARD",
	}

	// Настройка моков - симулируем отсутствие заказа
	s.orderRepository.On("GetOrder", mock.Anything, orderUUID).
		Return(nil, model.ErrOrderNotFound)

	// Вызов метода
	result, err := s.service.PayOrder(context.Background(), incomingOrder)

	// Проверка результата
	s.Require().Empty(result)
	s.Require().Error(err)
	s.Require().ErrorIs(err, model.ErrOrderNotFound)

	s.orderRepository.AssertExpectations(s.T())
}

func (s *ServiceSuite) TestPayOrder_PaymentError() {
	// Тестовые данные
	orderUUID := uuid.New()
	userUUID := uuid.New()
	partUUID := uuid.New()

	incomingOrder := &model.Order{
		OrderUUID:     orderUUID,
		PaymentMethod: "CARD",
	}

	dbOrder := &model.Order{
		OrderUUID:  orderUUID,
		UserUUID:   userUUID,
		PartUUIDs:  []uuid.UUID{partUUID},
		TotalPrice: 100.0,
		Status:     model.StatusPendingPayment,
	}

	// Настройка моков - успешное получение заказа, но ошибка при создании платежа
	s.orderRepository.On("GetOrder", mock.Anything, orderUUID).
		Return(dbOrder, nil)

	orderWithPaymentMethod := &model.Order{
		OrderUUID:     orderUUID,
		UserUUID:      userUUID,
		PartUUIDs:     []uuid.UUID{partUUID},
		TotalPrice:    100.0,
		Status:        model.StatusPendingPayment,
		PaymentMethod: "CARD",
	}

	s.paymentClient.On("CreatePayment", mock.Anything, orderWithPaymentMethod).
		Return(nil, model.ErrPaid)

	// Вызов метода
	result, err := s.service.PayOrder(context.Background(), incomingOrder)

	// Проверка результата
	s.Require().Empty(result)
	s.Require().Error(err)
	s.Require().ErrorIs(err, model.ErrPaid)

	s.orderRepository.AssertExpectations(s.T())
	s.paymentClient.AssertExpectations(s.T())
}

func (s *ServiceSuite) TestPayOrder_UpdateError() {
	// Тестовые данные
	orderUUID := uuid.New()
	userUUID := uuid.New()
	partUUID := uuid.New()
	transactionUUID := uuid.New()

	incomingOrder := &model.Order{
		OrderUUID:     orderUUID,
		PaymentMethod: "CARD",
	}

	dbOrder := &model.Order{
		OrderUUID:  orderUUID,
		UserUUID:   userUUID,
		PartUUIDs:  []uuid.UUID{partUUID},
		TotalPrice: 100.0,
		Status:     model.StatusPendingPayment,
	}

	paidOrder := &model.Order{
		OrderUUID:       orderUUID,
		UserUUID:        userUUID,
		PartUUIDs:       []uuid.UUID{partUUID},
		TotalPrice:      100.0,
		TransactionUUID: transactionUUID,
		PaymentMethod:   "CARD",
		Status:          model.StatusPaid,
	}

	// Настройка моков - успешное получение заказа и создание платежа, но ошибка при обновлении
	s.orderRepository.On("GetOrder", mock.Anything, orderUUID).
		Return(dbOrder, nil)

	orderWithPaymentMethod := &model.Order{
		OrderUUID:     orderUUID,
		UserUUID:      userUUID,
		PartUUIDs:     []uuid.UUID{partUUID},
		TotalPrice:    100.0,
		Status:        model.StatusPendingPayment,
		PaymentMethod: "CARD",
	}

	s.paymentClient.On("CreatePayment", mock.Anything, orderWithPaymentMethod).
		Return(paidOrder, nil)
	s.orderRepository.On("UpdateOrderIfStatus", mock.Anything, paidOrder, model.StatusPendingPayment).
		Return(nil, model.ErrOrderNotFound)
	// Деньги списаны, а заказ не обновлён: платёж должен быть возвращён
	s.paymentClient.On("RefundPayment", mock.Anything, paidOrder).Return(nil).Once()

	// Вызов метода
	result, err := s.service.PayOrder(context.Background(), incomingOrder)

	// Проверка результата
	s.Require().Empty(result)
	s.Require().Error(err)
	s.Require().ErrorIs(err, model.ErrOrderNotFound)

	s.orderRepository.AssertExpectations(s.T())
	s.paymentClient.AssertExpectations(s.T())
}

func (s *ServiceSuite) TestPayOrder_PaymentMethodPropagation() {
	// Тестовые данные
	orderUUID := uuid.New()
	userUUID := uuid.New()
	partUUID := uuid.New()
	transactionUUID := uuid.New()

	// Входящий запрос с различными методами оплаты
	testCases := []struct {
		name          string
		paymentMethod string
	}{
		{"Credit Card", "CREDIT_CARD"},
		{"SBP", "SBP"},
		{"Investor Money", "INVESTOR_MONEY"},
		{"Card", "CARD"},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			incomingOrder := &model.Order{
				OrderUUID:     orderUUID,
				PaymentMethod: tc.paymentMethod,
			}

			dbOrder := &model.Order{
				OrderUUID:  orderUUID,
				UserUUID:   userUUID,
				PartUUIDs:  []uuid.UUID{partUUID},
				TotalPrice: 100.0,
				Status:     model.StatusPendingPayment,
			}

			paidOrder := &model.Order{
				OrderUUID:       orderUUID,
				UserUUID:        userUUID,
				PartUUIDs:       []uuid.UUID{partUUID},
				TotalPrice:      100.0,
				TransactionUUID: transactionUUID,
				PaymentMethod:   tc.paymentMethod,
				Status:          model.StatusPaid,
			}

			// Настройка моков
			s.orderRepository.On("GetOrder", mock.Anything, orderUUID).
				Return(dbOrder, nil)

			orderWithPaymentMethod := &model.Order{
				OrderUUID:     orderUUID,
				UserUUID:      userUUID,
				PartUUIDs:     []uuid.UUID{partUUID},
				TotalPrice:    100.0,
				Status:        model.StatusPendingPayment,
				PaymentMethod: tc.paymentMethod,
			}

			s.paymentClient.On("CreatePayment", mock.Anything, orderWithPaymentMethod).
				Return(paidOrder, nil)
			s.orderRepository.On("UpdateOrderIfStatus", mock.Anything, paidOrder, model.StatusPendingPayment).
				Return(paidOrder, nil)

			// Вызов метода
			result, err := s.service.PayOrder(context.Background(), incomingOrder)

			// Проверка результата
			s.Require().NoError(err)
			s.Require().NotNil(result)
			s.Require().Equal(tc.paymentMethod, result.PaymentMethod)
			s.Require().Equal(model.StatusPaid, result.Status)

			// Очищаем моки для следующего теста
			s.orderRepository.ExpectedCalls = nil
			s.paymentClient.ExpectedCalls = nil
		})
	}
}

func (s *ServiceSuite) TestPayOrder_EventPaymentMethodPropagation() {
	// Тестовые данные
	orderUUID := uuid.New()
	userUUID := uuid.New()
	partUUID := uuid.New()
	transactionUUID := uuid.New()

	incomingOrder := &model.Order{
		OrderUUID:     orderUUID,
		PaymentMethod: "SBP",
	}

	dbOrder := &model.Order{
		OrderUUID:  orderUUID,
		UserUUID:   userUUID,
		PartUUIDs:  []uuid.UUID{partUUID},
		TotalPrice: 100.0,
		Status:     model.StatusPendingPayment,
	}

	paidOrder := &model.Order{
		OrderUUID:       orderUUID,
		UserUUID:        userUUID,
		PartUUIDs:       []uuid.UUID{partUUID},
		TotalPrice:      100.0,
		TransactionUUID: transactionUUID,
		PaymentMethod:   "SBP",
		Status:          model.StatusPaid,
	}

	// Настройка моков
	s.orderRepository.On("GetOrder", mock.Anything, orderUUID).
		Return(dbOrder, nil)

	orderWithPaymentMethod := &model.Order{
		OrderUUID:     orderUUID,
		UserUUID:      userUUID,
		PartUUIDs:     []uuid.UUID{partUUID},
		TotalPrice:    100.0,
		Status:        model.StatusPendingPayment,
		PaymentMethod: "SBP",
	}

	s.paymentClient.On("CreatePayment", mock.Anything, orderWithPaymentMethod).
		Return(paidOrder, nil)
	s.orderRepository.On("UpdateOrderIfStatus", mock.Anything, paidOrder, model.StatusPendingPayment).
		Return(paidOrder, nil)

	// Вызов метода
	result, err := s.service.PayOrder(context.Background(), incomingOrder)

	// Проверка результата
	s.Require().NoError(err)
	s.Require().NotNil(result)

	// Проверяем что событие было отправлено с правильным PaymentMethod
	lastEvent := s.orderPaidProducer.GetLastEvent()
	s.Require().NotNil(lastEvent)
	s.Require().Equal("SBP", lastEvent.PaymentMethod)
	s.Require().Equal(orderUUID, lastEvent.OrderUUID)
	s.Require().Equal(userUUID, lastEvent.UserUUID)
	s.Require().Equal(transactionUUID, lastEvent.TransactionUUID)

	s.orderRepository.AssertExpectations(s.T())
	s.paymentClient.AssertExpectations(s.T())
}

// Регрессия: оплатить можно только заказ в статусе PENDING_PAYMENT. Раньше оплата оплаченного,
// отменённого или уже собранного заказа создавала новый платёж и повторное событие order.paid.
func (s *ServiceSuite) TestPayOrder_NotPendingPaymentIsRejected() {
	statuses := []model.OrderStatus{model.StatusPaid, model.StatusCancelled, model.StatusAssembled}

	for _, status := range statuses {
		s.Run(string(status), func() {
			s.SetupTest()
			s.orderPaidProducer.lastEvent = nil

			orderUUID := uuid.New()
			s.orderRepository.On("GetOrder", mock.Anything, orderUUID).
				Return(&model.Order{
					OrderUUID: orderUUID,
					UserUUID:  uuid.New(),
					Status:    status,
				}, nil)
			// CreatePayment и UpdateOrder не мокаются: любой их вызов уронит тест

			result, err := s.service.PayOrder(context.Background(), &model.Order{
				OrderUUID:     orderUUID,
				PaymentMethod: "CARD",
			})

			s.Require().Nil(result)
			s.Require().ErrorIs(err, model.ErrOrderNotPayable)
			s.Require().Nil(s.orderPaidProducer.lastEvent, "событие order.paid публиковаться не должно")
		})
	}
}

// Параллельный запрос успел изменить статус между проверкой и записью:
// заказ не должен быть оплачен повторно, а событие order.paid не должно уйти второй раз.
func (s *ServiceSuite) TestPayOrder_StatusChangedConcurrently() {
	orderUUID := uuid.New()
	userUUID := uuid.New()

	dbOrder := &model.Order{
		OrderUUID:  orderUUID,
		UserUUID:   userUUID,
		TotalPrice: 100.0,
		Status:     model.StatusPendingPayment,
	}
	paidOrder := &model.Order{
		OrderUUID:       orderUUID,
		UserUUID:        userUUID,
		TotalPrice:      100.0,
		TransactionUUID: uuid.New(),
		PaymentMethod:   "SBP",
		Status:          model.StatusPaid,
	}

	s.orderRepository.On("GetOrder", mock.Anything, orderUUID).Return(dbOrder, nil)
	s.paymentClient.On("CreatePayment", mock.Anything, mock.AnythingOfType("*model.Order")).Return(paidOrder, nil)
	s.orderRepository.On("UpdateOrderIfStatus", mock.Anything, paidOrder, model.StatusPendingPayment).
		Return(nil, model.ErrOrderStatusChanged)

	result, err := s.service.PayOrder(context.Background(), &model.Order{OrderUUID: orderUUID, PaymentMethod: "SBP"})

	s.Require().Nil(result)
	s.Require().ErrorIs(err, model.ErrOrderNotPayable)
	s.Require().Nil(s.orderPaidProducer.GetLastEvent(), "событие не должно публиковаться при конфликте")
}

func (s *ServiceSuite) payScenario() (orderUUID uuid.UUID, paidOrder *model.Order) {
	orderUUID = uuid.New()
	dbOrder := &model.Order{OrderUUID: orderUUID, UserUUID: uuid.New(), TotalPrice: 100, Status: model.StatusPendingPayment}
	paidOrder = &model.Order{
		OrderUUID: orderUUID, UserUUID: dbOrder.UserUUID, TotalPrice: 100,
		TransactionUUID: uuid.New(), PaymentMethod: "CARD", Status: model.StatusPaid,
	}
	s.orderRepository.On("GetOrder", mock.Anything, orderUUID).Return(dbOrder, nil)
	s.paymentClient.On("CreatePayment", mock.Anything, mock.AnythingOfType("*model.Order")).Return(paidOrder, nil)
	return orderUUID, paidOrder
}

// Ошибка возврата не должна подменять исходную ошибку клиенту: возврат только логируется.
func (s *ServiceSuite) TestPayOrder_RefundFailureDoesNotMaskOriginalError() {
	orderUUID, paidOrder := s.payScenario()
	s.orderRepository.On("UpdateOrderIfStatus", mock.Anything, paidOrder, model.StatusPendingPayment).
		Return(nil, model.ErrFailedToUpdateOrder)
	s.paymentClient.On("RefundPayment", mock.Anything, paidOrder).Return(model.ErrPaymentClient).Once()

	_, err := s.service.PayOrder(context.Background(), &model.Order{OrderUUID: orderUUID, PaymentMethod: "CARD"})

	s.Require().ErrorIs(err, model.ErrFailedToUpdateOrder)
	s.Require().NotErrorIs(err, model.ErrPaymentClient)
}

// При конфликте статуса деньги принадлежат победившему запросу: возврат не вызывается
// (мок создан через NewPaymentClient, незапланированный вызов RefundPayment провалил бы тест).
func (s *ServiceSuite) TestPayOrder_ConflictDoesNotRefund() {
	orderUUID, paidOrder := s.payScenario()
	s.orderRepository.On("UpdateOrderIfStatus", mock.Anything, paidOrder, model.StatusPendingPayment).
		Return(nil, model.ErrOrderStatusChanged)

	_, err := s.service.PayOrder(context.Background(), &model.Order{OrderUUID: orderUUID, PaymentMethod: "CARD"})

	s.Require().ErrorIs(err, model.ErrOrderNotPayable)
	s.paymentClient.AssertNotCalled(s.T(), "RefundPayment", mock.Anything, mock.Anything)
}
