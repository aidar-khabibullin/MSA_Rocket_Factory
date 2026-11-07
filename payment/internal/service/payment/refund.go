package payment

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/aidar-khabibullin/MSA_Rocket_Factory/payment/internal/model"
	"github.com/aidar-khabibullin/MSA_Rocket_Factory/platform/pkg/logger"
	"github.com/aidar-khabibullin/MSA_Rocket_Factory/platform/pkg/tracing"
)

// Refund возвращает платёж по заказу. Повторный вызов для той же транзакции возвращает тот же refund_uuid.
func (s *service) Refund(ctx context.Context, orderUUID, transactionUUID uuid.UUID) (uuid.UUID, error) {
	_, span := tracing.StartSpan(ctx, "payment.refund",
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(
			attribute.String("payment.order_uuid", orderUUID.String()),
			attribute.String("payment.transaction_uuid", transactionUUID.String()),
		),
	)
	defer tracing.EndSpanWithStatus(span, nil)

	s.mu.Lock()
	defer s.mu.Unlock()

	rec, ok := s.records[orderUUID]
	if !ok {
		return uuid.Nil, fmt.Errorf("order %s: %w", orderUUID, model.ErrPaymentNotFound)
	}
	if rec.transactionUUID != transactionUUID {
		return uuid.Nil, fmt.Errorf("order %s: %w", orderUUID, model.ErrTransactionMismatch)
	}
	if rec.refundUUID != uuid.Nil {
		return rec.refundUUID, nil
	}

	rec.refundUUID = uuid.New()
	logger.Info(ctx, "Payment refunded",
		zap.String("order_uuid", orderUUID.String()),
		zap.String("transaction_uuid", transactionUUID.String()),
		zap.String("refund_uuid", rec.refundUUID.String()),
	)
	return rec.refundUUID, nil
}
