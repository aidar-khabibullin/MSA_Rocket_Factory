package payment

import (
	"context"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/aidar-khabibullin/MSA_Rocket_Factory/payment/internal/model"
	"github.com/aidar-khabibullin/MSA_Rocket_Factory/platform/pkg/logger"
	"github.com/aidar-khabibullin/MSA_Rocket_Factory/platform/pkg/tracing"
)

func (s *service) Pay(ctx context.Context, order model.Order) (uuid.UUID, error) {
	// Создаем спан для операции оплаты
	ctx, span := tracing.StartSpan(ctx, "payment.process", trace.WithSpanKind(trace.SpanKindInternal))
	defer tracing.EndSpanWithStatus(span, nil)

	// Добавляем атрибуты к спану
	span.SetAttributes(
		attribute.String("payment.order_uuid", order.OrderUuid.String()),
		attribute.String("payment.user_uuid", order.UserUuid.String()),
		attribute.String("payment.method", order.PaymentMethod),
	)

	// Платёж по заказу идемпотентен: повторный вызов (повтор запроса клиента, гонка двух запросов)
	// возвращает уже созданную транзакцию и не списывает деньги второй раз.
	transactionUuid, created := s.transactionFor(order.OrderUuid)
	if !created {
		span.AddEvent("idempotent replay: existing transaction returned")
		logger.Info(ctx, "Payment replay: existing transaction returned",
			zap.String("order_uuid", order.OrderUuid.String()),
			zap.String("transaction_uuid", transactionUuid.String()),
		)
		span.SetStatus(codes.Ok, "payment replayed")
		return transactionUuid, nil
	}
	span.AddEvent("transaction created")

	// Добавляем UUID транзакции к спану
	span.SetAttributes(attribute.String("payment.transaction_uuid", transactionUuid.String()))
	span.AddEvent("payment processed successfully", trace.WithAttributes(
		attribute.String("transaction_uuid", transactionUuid.String()),
	))

	logger.Info(ctx, "Payment success",
		zap.String("order_uuid", order.OrderUuid.String()),
		zap.String("transaction_uuid", transactionUuid.String()),
	)

	span.SetStatus(codes.Ok, "payment processed successfully")
	return transactionUuid, nil
}

// transactionFor возвращает транзакцию по заказу. created=false, если платёж уже существует и не возвращён.
func (s *service) transactionFor(orderUUID uuid.UUID) (transaction uuid.UUID, created bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if rec, ok := s.records[orderUUID]; ok && rec.refundUUID == uuid.Nil {
		return rec.transactionUUID, false
	}

	// Платежа нет или он возвращён: создаём новую транзакцию (повторная оплата после возврата).
	transaction = uuid.New()
	s.records[orderUUID] = &record{transactionUUID: transaction}
	return transaction, true
}
