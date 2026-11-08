package v1

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/aidar-khabibullin/MSA_Rocket_Factory/payment/internal/model"
	paymentV1 "github.com/aidar-khabibullin/MSA_Rocket_Factory/shared/pkg/proto/payment/v1"
)

func TestRefundOrder_Success(t *testing.T) {
	svc := new(MockPaymentService)
	orderUUID, txUUID, refundUUID := uuid.New(), uuid.New(), uuid.New()
	svc.On("Refund", mock.Anything, orderUUID, txUUID).Return(refundUUID, nil)

	resp, err := NewAPI(svc).RefundOrder(context.Background(), &paymentV1.RefundOrderRequest{
		OrderUuid: orderUUID.String(), TransactionUuid: txUUID.String(),
	})

	require.NoError(t, err)
	require.Equal(t, refundUUID.String(), resp.GetRefundUuid())
}

func TestRefundOrder_ErrorMapping(t *testing.T) {
	valid := uuid.New().String()
	tests := []struct {
		name     string
		order    string
		tx       string
		svcErr   error
		wantCode codes.Code
	}{
		{"невалидный order_uuid", "x", valid, nil, codes.InvalidArgument},
		{"невалидный transaction_uuid", valid, "x", nil, codes.InvalidArgument},
		{"платёж не найден", valid, valid, model.ErrPaymentNotFound, codes.NotFound},
		{"транзакция не совпадает", valid, valid, model.ErrTransactionMismatch, codes.FailedPrecondition},
		{"прочая ошибка", valid, valid, errors.New("boom"), codes.Internal},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := new(MockPaymentService)
			if tt.svcErr != nil {
				svc.On("Refund", mock.Anything, mock.Anything, mock.Anything).Return(nil, tt.svcErr)
			}

			_, err := NewAPI(svc).RefundOrder(context.Background(), &paymentV1.RefundOrderRequest{
				OrderUuid: tt.order, TransactionUuid: tt.tx,
			})

			require.Equal(t, tt.wantCode, status.Code(err))
		})
	}
}
