package v1

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/aidar-khabibullin/MSA_Rocket_Factory/payment/internal/model"
	"github.com/aidar-khabibullin/MSA_Rocket_Factory/platform/pkg/tracing"
	paymentV1 "github.com/aidar-khabibullin/MSA_Rocket_Factory/shared/pkg/proto/payment/v1"
)

func (a *api) RefundOrder(ctx context.Context, req *paymentV1.RefundOrderRequest) (*paymentV1.RefundOrderResponse, error) {
	ctx, span := tracing.StartSpan(ctx, "payment.refund_order",
		trace.WithAttributes(attribute.String("order_uuid", req.GetOrderUuid())),
	)
	defer span.End()

	orderUUID, err := uuid.Parse(req.GetOrderUuid())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid order_uuid")
	}
	transactionUUID, err := uuid.Parse(req.GetTransactionUuid())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid transaction_uuid")
	}

	refundUUID, err := a.paymentService.Refund(ctx, orderUUID, transactionUUID)
	switch {
	case errors.Is(err, model.ErrPaymentNotFound):
		return nil, status.Error(codes.NotFound, "payment for this order not found")
	case errors.Is(err, model.ErrTransactionMismatch):
		return nil, status.Error(codes.FailedPrecondition, "transaction does not match the order payment")
	case err != nil:
		span.RecordError(err)
		return nil, status.Error(codes.Internal, "refund failed")
	}

	return &paymentV1.RefundOrderResponse{RefundUuid: refundUUID.String()}, nil
}
