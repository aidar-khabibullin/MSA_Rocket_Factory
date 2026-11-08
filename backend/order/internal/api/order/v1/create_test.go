package v1

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/aidar-khabibullin/MSA_Rocket_Factory/order/internal/model"
	"github.com/aidar-khabibullin/MSA_Rocket_Factory/order/internal/service/mocks"
	"github.com/aidar-khabibullin/MSA_Rocket_Factory/platform/pkg/logger"
	grpcAuth "github.com/aidar-khabibullin/MSA_Rocket_Factory/platform/pkg/middleware/grpc"
	orderV1 "github.com/aidar-khabibullin/MSA_Rocket_Factory/shared/pkg/openapi/order/v1"
	iamV1 "github.com/aidar-khabibullin/MSA_Rocket_Factory/shared/pkg/proto/iam/v1"
)

// authedCtx возвращает контекст с аутентифицированным пользователем, как это делает AuthMiddleware
func authedCtx(userUUID uuid.UUID) context.Context {
	return context.WithValue(context.Background(), grpcAuth.GetUserContextKey(), &iamV1.User{Uuid: userUUID.String()})
}

// expectOwnedOrder настраивает мок так, что заказ orderUUID существует и принадлежит новому пользователю,
// и возвращает контекст этого пользователя (как после AuthMiddleware). Используется в тестах pay/cancel,
// где хэндлер сначала проверяет владельца через GetOrder.
func expectOwnedOrder(m *mocks.OrderService, orderUUID uuid.UUID) context.Context {
	owner := uuid.New()
	m.EXPECT().
		GetOrder(mock.Anything, orderUUID).
		Return(&model.Order{OrderUUID: orderUUID, UserUUID: owner, Status: model.StatusPendingPayment}, nil).
		Once()

	return authedCtx(owner)
}

func init() {
	// Инициализируем logger для тестов
	_ = logger.InitSimple("error", false)
}

func TestAPI_CreateOrder_Success(t *testing.T) {
	// Arrange
	mockService := mocks.NewOrderService(t)
	api := NewAPI(mockService)

	userUUID := uuid.New()
	partUUID1 := uuid.New()
	partUUID2 := uuid.New()
	orderUUID := uuid.New()

	req := &orderV1.CreateOrderRequest{
		UserUUID:  orderV1.UserUUID(userUUID),
		PartUuids: []uuid.UUID{partUUID1, partUUID2},
	}

	expectedOrder := &model.Order{
		OrderUUID:  orderUUID,
		UserUUID:   userUUID,
		PartUUIDs:  []uuid.UUID{partUUID1, partUUID2},
		TotalPrice: 1500.50,
		Status:     "PENDING_PAYMENT",
	}

	mockService.EXPECT().
		CreateOrder(mock.Anything, mock.MatchedBy(func(order *model.Order) bool {
			return order.UserUUID == userUUID &&
				len(order.PartUUIDs) == 2 &&
				order.PartUUIDs[0] == partUUID1 &&
				order.PartUUIDs[1] == partUUID2
		})).
		Return(expectedOrder, nil).
		Once()

	// Act
	res, err := api.CreateOrder(authedCtx(uuid.UUID(req.UserUUID)), req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, res)

	response, ok := res.(*orderV1.CreateOrderResponse)
	assert.True(t, ok)
	assert.Equal(t, orderV1.OrderUUID(orderUUID), response.OrderUUID)
	assert.True(t, response.TotalPrice.Set)
	assert.Equal(t, orderV1.TotalPrice(1500.50), response.TotalPrice.Value)

	mockService.AssertExpectations(t)
}

func TestAPI_CreateOrder_ServiceNil(t *testing.T) {
	// Arrange
	api := &api{orderService: nil}

	req := &orderV1.CreateOrderRequest{
		UserUUID:  orderV1.UserUUID(uuid.New()),
		PartUuids: []uuid.UUID{uuid.New()},
	}

	// Act
	res, err := api.CreateOrder(authedCtx(uuid.UUID(req.UserUUID)), req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, res)

	errorResponse, ok := res.(*orderV1.InternalServerError)
	assert.True(t, ok)
	assert.Equal(t, 500, errorResponse.Code)
	assert.Equal(t, "order service not available", errorResponse.Message)
}

func TestAPI_CreateOrder_PartsNotSpecified(t *testing.T) {
	// Arrange
	mockService := mocks.NewOrderService(t)
	api := NewAPI(mockService)

	req := &orderV1.CreateOrderRequest{
		UserUUID:  orderV1.UserUUID(uuid.New()),
		PartUuids: []uuid.UUID{},
	}

	mockService.EXPECT().
		CreateOrder(mock.Anything, mock.Anything).
		Return(nil, model.ErrPartsSpecified).
		Once()

	// Act
	res, err := api.CreateOrder(authedCtx(uuid.UUID(req.UserUUID)), req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, res)

	errorResponse, ok := res.(*orderV1.BadRequestError)
	assert.True(t, ok)
	assert.Equal(t, 400, errorResponse.Code)
	assert.Equal(t, "parts not specified", errorResponse.Message)

	mockService.AssertExpectations(t)
}

func TestAPI_CreateOrder_PartsNotFound(t *testing.T) {
	// Arrange
	mockService := mocks.NewOrderService(t)
	api := NewAPI(mockService)

	req := &orderV1.CreateOrderRequest{
		UserUUID:  orderV1.UserUUID(uuid.New()),
		PartUuids: []uuid.UUID{uuid.New()},
	}

	mockService.EXPECT().
		CreateOrder(mock.Anything, mock.Anything).
		Return(nil, model.ErrPartsListNotFound).
		Once()

	// Act
	res, err := api.CreateOrder(authedCtx(uuid.UUID(req.UserUUID)), req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, res)

	errorResponse, ok := res.(*orderV1.NotFoundError)
	assert.True(t, ok)
	assert.Equal(t, 404, errorResponse.Code)
	assert.Equal(t, "parts not found", errorResponse.Message)

	mockService.AssertExpectations(t)
}

func TestAPI_CreateOrder_ConvertFromClientError(t *testing.T) {
	// Arrange
	mockService := mocks.NewOrderService(t)
	api := NewAPI(mockService)

	req := &orderV1.CreateOrderRequest{
		UserUUID:  orderV1.UserUUID(uuid.New()),
		PartUuids: []uuid.UUID{uuid.New()},
	}

	mockService.EXPECT().
		CreateOrder(mock.Anything, mock.Anything).
		Return(nil, model.ErrConvertFromClient).
		Once()

	// Act
	res, err := api.CreateOrder(authedCtx(uuid.UUID(req.UserUUID)), req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, res)

	errorResponse, ok := res.(*orderV1.InternalServerError)
	assert.True(t, ok)
	assert.Equal(t, 500, errorResponse.Code)
	assert.Equal(t, "can't parse part", errorResponse.Message)

	mockService.AssertExpectations(t)
}

func TestAPI_CreateOrder_InventoryClientError(t *testing.T) {
	// Arrange
	mockService := mocks.NewOrderService(t)
	api := NewAPI(mockService)

	req := &orderV1.CreateOrderRequest{
		UserUUID:  orderV1.UserUUID(uuid.New()),
		PartUuids: []uuid.UUID{uuid.New()},
	}

	mockService.EXPECT().
		CreateOrder(mock.Anything, mock.Anything).
		Return(nil, model.ErrInventoryClient).
		Once()

	// Act
	res, err := api.CreateOrder(authedCtx(uuid.UUID(req.UserUUID)), req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, res)

	errorResponse, ok := res.(*orderV1.InternalServerError)
	assert.True(t, ok)
	assert.Equal(t, 424, errorResponse.Code) // HTTP_FAILED_DEPENDENCY
	assert.Equal(t, "inventory client error", errorResponse.Message)

	mockService.AssertExpectations(t)
}

func TestAPI_CreateOrder_TableDriven(t *testing.T) {
	tests := []struct {
		name               string
		setupMock          func(*mocks.OrderService)
		request            *orderV1.CreateOrderRequest
		expectedStatusCode int
		expectedMessage    string
		expectError        bool
	}{
		{
			name: "успешное создание с одной деталью",
			setupMock: func(m *mocks.OrderService) {
				m.EXPECT().
					CreateOrder(mock.Anything, mock.Anything).
					Return(&model.Order{
						OrderUUID:  uuid.New(),
						UserUUID:   uuid.New(),
						TotalPrice: 100.0,
					}, nil).
					Once()
			},
			request: &orderV1.CreateOrderRequest{
				UserUUID:  orderV1.UserUUID(uuid.New()),
				PartUuids: []uuid.UUID{uuid.New()},
			},
			expectedStatusCode: 0, // Success response не имеет статус кода в структуре
			expectError:        false,
		},
		{
			name: "пустой список деталей",
			setupMock: func(m *mocks.OrderService) {
				m.EXPECT().
					CreateOrder(mock.Anything, mock.Anything).
					Return(nil, model.ErrPartsSpecified).
					Once()
			},
			request: &orderV1.CreateOrderRequest{
				UserUUID:  orderV1.UserUUID(uuid.New()),
				PartUuids: []uuid.UUID{},
			},
			expectedStatusCode: 400,
			expectedMessage:    "parts not specified",
			expectError:        false,
		},
		{
			name: "детали не найдены в inventory",
			setupMock: func(m *mocks.OrderService) {
				m.EXPECT().
					CreateOrder(mock.Anything, mock.Anything).
					Return(nil, model.ErrPartsListNotFound).
					Once()
			},
			request: &orderV1.CreateOrderRequest{
				UserUUID:  orderV1.UserUUID(uuid.New()),
				PartUuids: []uuid.UUID{uuid.New()},
			},
			expectedStatusCode: 404,
			expectedMessage:    "parts not found",
			expectError:        false,
		},
		{
			name: "недоступен inventory service",
			setupMock: func(m *mocks.OrderService) {
				m.EXPECT().
					CreateOrder(mock.Anything, mock.Anything).
					Return(nil, model.ErrInventoryClient).
					Once()
			},
			request: &orderV1.CreateOrderRequest{
				UserUUID:  orderV1.UserUUID(uuid.New()),
				PartUuids: []uuid.UUID{uuid.New()},
			},
			expectedStatusCode: 424,
			expectedMessage:    "inventory client error",
			expectError:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockService := mocks.NewOrderService(t)
			tt.setupMock(mockService)
			api := NewAPI(mockService)

			// Act
			res, err := api.CreateOrder(authedCtx(uuid.UUID(tt.request.UserUUID)), tt.request)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, res)

				// Проверяем тип ответа в зависимости от ожидаемого статус кода
				switch tt.expectedStatusCode {
				case 0: // Success
					_, ok := res.(*orderV1.CreateOrderResponse)
					assert.True(t, ok, "ожидался CreateOrderResponse")
				case 400:
					errorResponse, ok := res.(*orderV1.BadRequestError)
					assert.True(t, ok, "ожидался BadRequestError")
					assert.Equal(t, tt.expectedStatusCode, errorResponse.Code)
					assert.Equal(t, tt.expectedMessage, errorResponse.Message)
				case 404:
					errorResponse, ok := res.(*orderV1.NotFoundError)
					assert.True(t, ok, "ожидался NotFoundError")
					assert.Equal(t, tt.expectedStatusCode, errorResponse.Code)
					assert.Equal(t, tt.expectedMessage, errorResponse.Message)
				case 424, 500:
					errorResponse, ok := res.(*orderV1.InternalServerError)
					assert.True(t, ok, "ожидался InternalServerError")
					assert.Equal(t, tt.expectedStatusCode, errorResponse.Code)
					assert.Equal(t, tt.expectedMessage, errorResponse.Message)
				}
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestAPI_CreateOrder_NoAuthenticatedUser(t *testing.T) {
	mockService := mocks.NewOrderService(t) // сервис вызываться не должен
	api := NewAPI(mockService)

	res, err := api.CreateOrder(context.Background(), &orderV1.CreateOrderRequest{
		UserUUID:  orderV1.UserUUID(uuid.New()),
		PartUuids: []uuid.UUID{uuid.New()},
	})

	assert.NoError(t, err)
	unauthorized, ok := res.(*orderV1.UnauthorizedError)
	assert.True(t, ok, "ожидался UnauthorizedError, получен %T", res)
	assert.Equal(t, 401, unauthorized.Code)
}

// Регрессия: клиент не должен иметь возможности создать заказ от имени другого пользователя
func TestAPI_CreateOrder_ForeignUserUUIDIsForbidden(t *testing.T) {
	mockService := mocks.NewOrderService(t) // сервис вызываться не должен
	api := NewAPI(mockService)

	sessionUser := uuid.New()
	someoneElse := uuid.New()

	res, err := api.CreateOrder(authedCtx(sessionUser), &orderV1.CreateOrderRequest{
		UserUUID:  orderV1.UserUUID(someoneElse),
		PartUuids: []uuid.UUID{uuid.New()},
	})

	assert.NoError(t, err)
	forbidden, ok := res.(*orderV1.ForbiddenError)
	assert.True(t, ok, "ожидался ForbiddenError, получен %T", res)
	assert.Equal(t, 403, forbidden.Code)
}

func TestAPI_CreateOrder_OwnerComesFromSession(t *testing.T) {
	mockService := mocks.NewOrderService(t)
	api := NewAPI(mockService)

	sessionUser := uuid.New()
	partUUID := uuid.New()

	mockService.EXPECT().
		CreateOrder(mock.Anything, mock.MatchedBy(func(order *model.Order) bool {
			return order.UserUUID == sessionUser
		})).
		Return(&model.Order{OrderUUID: uuid.New(), UserUUID: sessionUser, TotalPrice: 10}, nil).
		Once()

	// user_uuid в теле не передан (нулевой) - владелец берётся из сессии
	res, err := api.CreateOrder(authedCtx(sessionUser), &orderV1.CreateOrderRequest{
		PartUuids: []uuid.UUID{partUUID},
	})

	assert.NoError(t, err)
	_, ok := res.(*orderV1.CreateOrderResponse)
	assert.True(t, ok, "ожидался CreateOrderResponse, получен %T", res)
}

func TestAPI_CreateOrder_InvalidSessionUserUUID(t *testing.T) {
	mockService := mocks.NewOrderService(t) // сервис вызываться не должен
	api := NewAPI(mockService)

	ctx := context.WithValue(context.Background(), grpcAuth.GetUserContextKey(), &iamV1.User{Uuid: "not-a-uuid"})

	res, err := api.CreateOrder(ctx, &orderV1.CreateOrderRequest{PartUuids: []uuid.UUID{uuid.New()}})

	assert.NoError(t, err)
	_, ok := res.(*orderV1.InternalServerError)
	assert.True(t, ok, "ожидался InternalServerError, получен %T", res)
}
