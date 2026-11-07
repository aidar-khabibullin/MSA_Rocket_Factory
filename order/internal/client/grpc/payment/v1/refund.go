package v1

import (
	"context"

	"github.com/aidar-khabibullin/MSA_Rocket_Factory/order/internal/model"
	grpcAuth "github.com/aidar-khabibullin/MSA_Rocket_Factory/platform/pkg/middleware/grpc"
	generaredPaymentV1 "github.com/aidar-khabibullin/MSA_Rocket_Factory/shared/pkg/proto/payment/v1"
)

// RefundPayment возвращает платёж заказа. Вызов идемпотентен на стороне payment.
func (p paymentClient) RefundPayment(ctx context.Context, order *model.Order) error {
	ctx = grpcAuth.ForwardSessionUUIDToGRPC(ctx)

	_, err := p.generatedClient.RefundOrder(ctx, &generaredPaymentV1.RefundOrderRequest{
		OrderUuid:       order.OrderUUID.String(),
		TransactionUuid: order.TransactionUUID.String(),
	})
	if err != nil {
		return model.ErrPaymentClient
	}

	return nil
}
