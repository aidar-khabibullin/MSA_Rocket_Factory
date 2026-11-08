package order_test

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/aidar-khabibullin/MSA_Rocket_Factory/order/internal/model"
)

// resetCalls очищает историю вызовов общих моков: AssertNotCalled иначе видит вызовы предыдущих тестов набора.
func (s *ServiceSuite) resetCalls() {
	s.orderRepository.Calls = nil
	s.inventoryClient.Calls = nil
	s.paymentClient.Calls = nil
}

// stockScenario готовит заказ из двух деталей с ответом ListParts; резервные моки настраивает сам тест.
func (s *ServiceSuite) stockScenario() (*model.Order, []uuid.UUID) {
	s.resetCalls()
	partA, partB := uuid.New(), uuid.New()
	parts := []model.Part{{PartUUID: partA, Price: 100}, {PartUUID: partB, Price: 200}}
	s.inventoryClient.ExpectedCalls = nil
	s.inventoryClient.On("ListParts", mock.Anything, mock.AnythingOfType("*model.Filter")).Return(&parts, nil)

	return &model.Order{UserUUID: uuid.New(), PartUUIDs: []uuid.UUID{partA, partB}}, []uuid.UUID{partA, partB}
}

func (s *ServiceSuite) TestCreateOrder_ReservesPartsUnderTheSameOrderUUIDThatIsSaved() {
	order, partUUIDs := s.stockScenario()

	var reservedFor uuid.UUID
	s.inventoryClient.On("ReserveParts", mock.Anything, mock.AnythingOfType("uuid.UUID"), partUUIDs).
		Run(func(args mock.Arguments) { reservedFor = args.Get(1).(uuid.UUID) }).Return(nil).Once()
	s.orderRepository.On("CreateOrder", mock.Anything, mock.MatchedBy(func(o *model.Order) bool {
		return o.OrderUUID != uuid.Nil && o.OrderUUID == reservedFor
	})).Return(order, nil).Once()

	_, err := s.service.CreateOrder(context.Background(), order)

	s.Require().NoError(err)
	s.NotEqual(uuid.Nil, reservedFor, "резерв делается под UUID, с которым заказ потом сохраняется")
}

func (s *ServiceSuite) TestCreateOrder_OutOfStockIsRejectedAndNothingIsSaved() {
	order, _ := s.stockScenario()
	s.inventoryClient.On("ReserveParts", mock.Anything, mock.Anything, mock.Anything).Return(model.ErrPartsOutOfStock).Once()

	result, err := s.service.CreateOrder(context.Background(), order)

	s.Require().Nil(result)
	s.Require().ErrorIs(err, model.ErrPartsOutOfStock)
	s.orderRepository.AssertNotCalled(s.T(), "CreateOrder", mock.Anything, mock.Anything)
}

func (s *ServiceSuite) TestCreateOrder_SaveFailureReleasesReservation() {
	order, _ := s.stockScenario()
	s.inventoryClient.On("ReserveParts", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
	s.orderRepository.On("CreateOrder", mock.Anything, mock.Anything).Return(nil, model.ErrFailedToInsertOrder).Once()
	s.inventoryClient.On("ReleaseParts", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(nil).Once()

	_, err := s.service.CreateOrder(context.Background(), order)

	s.Require().ErrorIs(err, model.ErrFailedToInsertOrder)
	s.inventoryClient.AssertCalled(s.T(), "ReleaseParts", mock.Anything, mock.Anything)
}

func (s *ServiceSuite) TestCancelOrder_ReleasesReservation() {
	s.resetCalls()
	orderUUID := uuid.New()
	pending := &model.Order{OrderUUID: orderUUID, Status: model.StatusPendingPayment}
	s.orderRepository.On("GetOrder", mock.Anything, orderUUID).Return(pending, nil)
	s.orderRepository.On("UpdateOrder", mock.Anything, mock.Anything).Return(&model.Order{OrderUUID: orderUUID, Status: model.StatusCancelled}, nil)
	s.inventoryClient.ExpectedCalls = nil
	s.inventoryClient.On("ReleaseParts", mock.Anything, orderUUID).Return(nil).Once()

	result, err := s.service.CancelOrder(context.Background(), &model.Order{OrderUUID: orderUUID})

	s.Require().NoError(err)
	s.Equal(model.StatusCancelled, result.Status)
	s.inventoryClient.AssertCalled(s.T(), "ReleaseParts", mock.Anything, orderUUID)
}

// Отмена уже состоялась: сбой снятия резерва не должен превращать её в ошибку для клиента.
func (s *ServiceSuite) TestCancelOrder_ReleaseFailureDoesNotFailCancellation() {
	orderUUID := uuid.New()
	s.orderRepository.On("GetOrder", mock.Anything, orderUUID).Return(&model.Order{OrderUUID: orderUUID, Status: model.StatusPendingPayment}, nil)
	s.orderRepository.On("UpdateOrder", mock.Anything, mock.Anything).Return(&model.Order{OrderUUID: orderUUID, Status: model.StatusCancelled}, nil)
	s.inventoryClient.ExpectedCalls = nil
	s.inventoryClient.On("ReleaseParts", mock.Anything, orderUUID).Return(errors.New("inventory down")).Once()

	result, err := s.service.CancelOrder(context.Background(), &model.Order{OrderUUID: orderUUID})

	s.Require().NoError(err)
	s.Equal(model.StatusCancelled, result.Status)
}

func (s *ServiceSuite) TestCancelOrder_AlreadyPaidDoesNotTouchStock() {
	s.resetCalls()
	orderUUID := uuid.New()
	s.orderRepository.On("GetOrder", mock.Anything, orderUUID).Return(&model.Order{OrderUUID: orderUUID, Status: model.StatusPaid}, nil)
	s.inventoryClient.ExpectedCalls = nil

	_, err := s.service.CancelOrder(context.Background(), &model.Order{OrderUUID: orderUUID})

	s.Require().ErrorIs(err, model.ErrPaid)
	s.inventoryClient.AssertNotCalled(s.T(), "ReleaseParts", mock.Anything, mock.Anything)
}
