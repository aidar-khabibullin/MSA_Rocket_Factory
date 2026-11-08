package v1

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/aidar-khabibullin/MSA_Rocket_Factory/order/internal/model"
	"github.com/aidar-khabibullin/MSA_Rocket_Factory/order/internal/service/mocks"
	orderV1 "github.com/aidar-khabibullin/MSA_Rocket_Factory/shared/pkg/openapi/order/v1"
)

// foreignOrder настраивает мок: заказ существует, но принадлежит другому пользователю.
// Возвращает контекст "посторонней" сессии.
func foreignOrder(m *mocks.OrderService, orderUUID uuid.UUID) context.Context {
	m.EXPECT().
		GetOrder(mock.Anything, orderUUID).
		Return(&model.Order{OrderUUID: orderUUID, UserUUID: uuid.New(), Status: model.StatusPendingPayment}, nil).
		Once()

	return authedCtx(uuid.New())
}

// --- чужой заказ: ответ неотличим от "заказ не найден", действие не выполняется ---

func TestAPI_GetOrderByUUID_ForeignOrderIsNotFound(t *testing.T) {
	mockService := mocks.NewOrderService(t)
	api := NewAPI(mockService)
	orderUUID := uuid.New()

	res, err := api.GetOrderByUUID(foreignOrder(mockService, orderUUID), orderV1.GetOrderByUUIDParams{OrderUUID: orderUUID})

	assert.NoError(t, err)
	notFound, ok := res.(*orderV1.NotFoundError)
	assert.True(t, ok, "ожидался NotFoundError, получен %T", res)
	assert.Equal(t, 404, notFound.Code)
}

func TestAPI_PayOrder_ForeignOrderIsNotFound(t *testing.T) {
	mockService := mocks.NewOrderService(t) // PayOrder вызываться не должен
	api := NewAPI(mockService)
	orderUUID := uuid.New()

	res, err := api.PayOrder(
		foreignOrder(mockService, orderUUID),
		&orderV1.PayOrderRequest{PaymentMethod: orderV1.PaymentMethodCARD},
		orderV1.PayOrderParams{OrderUUID: orderUUID},
	)

	assert.NoError(t, err)
	notFound, ok := res.(*orderV1.NotFoundError)
	assert.True(t, ok, "ожидался NotFoundError, получен %T", res)
	assert.Equal(t, 404, notFound.Code)
}

func TestAPI_CancelOrder_ForeignOrderIsNotFound(t *testing.T) {
	mockService := mocks.NewOrderService(t) // CancelOrder вызываться не должен
	api := NewAPI(mockService)
	orderUUID := uuid.New()

	res, err := api.CancelOrder(foreignOrder(mockService, orderUUID), orderV1.CancelOrderParams{OrderUUID: orderUUID})

	assert.NoError(t, err)
	notFound, ok := res.(*orderV1.NotFoundError)
	assert.True(t, ok, "ожидался NotFoundError, получен %T", res)
	assert.Equal(t, 404, notFound.Code)
}

// --- нет аутентифицированного пользователя: сервис не вызывается вообще ---

func TestAPI_GetOrderByUUID_Unauthenticated(t *testing.T) {
	api := NewAPI(mocks.NewOrderService(t))

	res, err := api.GetOrderByUUID(context.Background(), orderV1.GetOrderByUUIDParams{OrderUUID: uuid.New()})

	assert.NoError(t, err)
	_, ok := res.(*orderV1.UnauthorizedError)
	assert.True(t, ok, "ожидался UnauthorizedError, получен %T", res)
}

func TestAPI_PayOrder_Unauthenticated(t *testing.T) {
	api := NewAPI(mocks.NewOrderService(t))

	res, err := api.PayOrder(
		context.Background(),
		&orderV1.PayOrderRequest{PaymentMethod: orderV1.PaymentMethodCARD},
		orderV1.PayOrderParams{OrderUUID: uuid.New()},
	)

	assert.NoError(t, err)
	_, ok := res.(*orderV1.UnauthorizedError)
	assert.True(t, ok, "ожидался UnauthorizedError, получен %T", res)
}

func TestAPI_CancelOrder_Unauthenticated(t *testing.T) {
	api := NewAPI(mocks.NewOrderService(t))

	res, err := api.CancelOrder(context.Background(), orderV1.CancelOrderParams{OrderUUID: uuid.New()})

	assert.NoError(t, err)
	_, ok := res.(*orderV1.UnauthorizedError)
	assert.True(t, ok, "ожидался UnauthorizedError, получен %T", res)
}

// --- оплата заказа в неподходящем статусе ---

func TestAPI_PayOrder_NotPayableIsConflict(t *testing.T) {
	mockService := mocks.NewOrderService(t)
	api := NewAPI(mockService)
	orderUUID := uuid.New()

	ctx := expectOwnedOrder(mockService, orderUUID)
	mockService.EXPECT().
		PayOrder(mock.Anything, mock.Anything).
		Return(nil, model.ErrOrderNotPayable).
		Once()

	res, err := api.PayOrder(
		ctx,
		&orderV1.PayOrderRequest{PaymentMethod: orderV1.PaymentMethodCARD},
		orderV1.PayOrderParams{OrderUUID: orderUUID},
	)

	assert.NoError(t, err)
	conflict, ok := res.(*orderV1.ConflictError)
	assert.True(t, ok, "ожидался ConflictError, получен %T", res)
	assert.Equal(t, 409, conflict.Code)
}
