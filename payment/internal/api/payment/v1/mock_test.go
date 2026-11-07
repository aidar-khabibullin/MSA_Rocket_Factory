package v1

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/aidar-khabibullin/MSA_Rocket_Factory/payment/internal/model"
)

// MockPaymentService мок для PaymentService
type MockPaymentService struct {
	mock.Mock
}

func (m *MockPaymentService) Pay(ctx context.Context, order model.Order) (uuid.UUID, error) {
	args := m.Called(ctx, order)
	if args.Get(0) == nil {
		return uuid.Nil, args.Error(1)
	}
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *MockPaymentService) Refund(ctx context.Context, orderUUID, transactionUUID uuid.UUID) (uuid.UUID, error) {
	args := m.Called(ctx, orderUUID, transactionUUID)
	if args.Get(0) == nil {
		return uuid.Nil, args.Error(1)
	}
	return args.Get(0).(uuid.UUID), args.Error(1)
}
