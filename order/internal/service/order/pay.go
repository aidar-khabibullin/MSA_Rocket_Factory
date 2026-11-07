package order

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/aidar-khabibullin/MSA_Rocket_Factory/order/internal/model"
	"github.com/aidar-khabibullin/MSA_Rocket_Factory/platform/pkg/logger"
	"github.com/aidar-khabibullin/MSA_Rocket_Factory/platform/pkg/tracing"
)

func (s service) PayOrder(ctx context.Context, order *model.Order) (*model.Order, error) {
	startTime := time.Now()

	// Создаем span для операции оплаты в сервисе
	ctx, span := tracing.StartSpan(ctx, "order.service.pay_order",
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(
			attribute.String("order_uuid", order.OrderUUID.String()),
			attribute.String("payment_method", order.PaymentMethod),
		),
	)
	defer tracing.EndSpanWithStatus(span, nil)

	// Получаем заказ по UUID
	span.AddEvent("fetching order from repository")
	dbOrder, err := s.orderRepository.GetOrder(ctx, order.OrderUUID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		span.End()
		return nil, fmt.Errorf("service: failed to get order from repository: %w", err)
	}
	span.AddEvent("order fetched successfully")

	// Оплатить можно только заказ, ожидающий оплаты: иначе возможны повторные платежи
	// и оплата отменённых заказов (с повторной публикацией события order.paid)
	if dbOrder.Status != model.StatusPendingPayment {
		err := fmt.Errorf("service: order %s has status %s: %w", dbOrder.OrderUUID, dbOrder.Status, model.ErrOrderNotPayable)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		span.End()
		return nil, err
	}

	// Устанавливаем метод оплаты из запроса
	dbOrder.PaymentMethod = order.PaymentMethod

	// Выполняем запрос к API для оплаты заказа
	span.AddEvent("processing payment via payment client")
	paidOrder, err := s.paymentClient.CreatePayment(ctx, dbOrder)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		span.End()
		return nil, fmt.Errorf("service: failed to create payment in payment client: %w", err)
	}
	span.AddEvent("payment processed successfully", trace.WithAttributes(
		attribute.String("transaction_uuid", paidOrder.TransactionUUID.String()),
	))

	// Обновляем заказ в хранилище
	span.AddEvent("updating order in repository")
	// Условный UPDATE: если параллельный запрос уже изменил статус, не публикуем повторное событие
	updatedOrder, err := s.orderRepository.UpdateOrderIfStatus(ctx, paidOrder, model.StatusPendingPayment)
	if errors.Is(err, model.ErrOrderStatusChanged) {
		// Параллельный запрос уже оплатил заказ: платёж по заказу идемпотентен, транзакция у обоих одна,
		// поэтому возвращать деньги нельзя, они принадлежат победившему запросу.
		err = fmt.Errorf("service: order %s changed status during payment: %w", dbOrder.OrderUUID, model.ErrOrderNotPayable)
	} else if err != nil {
		// Деньги списаны, а заказ не обновлён: компенсируем, иначе клиент заплатил бы за заказ в PENDING_PAYMENT.
		s.compensatePayment(ctx, paidOrder)
	}
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		span.End()
		return nil, fmt.Errorf("service: failed to update order in repository: %w", err)
	}

	// Отправляем событие OrderPaid (если producer доступен)
	if s.orderPaidProducer != nil {
		span.AddEvent("producing order paid event")
		event := model.OrderPaidEvent{
			EventUUID:       uuid.New(),
			OrderUUID:       updatedOrder.OrderUUID,
			UserUUID:        updatedOrder.UserUUID,
			PaymentMethod:   updatedOrder.PaymentMethod,
			TransactionUUID: updatedOrder.TransactionUUID,
		}

		err = s.orderPaidProducer.ProduceOrderPaid(ctx, event)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			span.End()
			return nil, fmt.Errorf("service: failed to produce OrderPaid event: %w", err)
		}
		span.AddEvent("order paid event produced successfully")
	}

	// Записываем метрики при успешной оплате заказа
	if s.metrics != nil {
		s.metrics.recordOrderPaid(ctx, float64(updatedOrder.TotalPrice), "USD") // По умолчанию USD
		s.metrics.recordOrderDuration(ctx, time.Since(startTime), "pay")
	}

	span.SetStatus(codes.Ok, "order paid successfully")
	return updatedOrder, nil
}

// compensatePayment возвращает платёж, не удавшийся довести до записи в заказе. Ошибка возврата не маскирует
// исходную: она логируется, а возврат повторяется оператором или повторной оплатой (идемпотентно).
func (s service) compensatePayment(ctx context.Context, paidOrder *model.Order) {
	if err := s.paymentClient.RefundPayment(ctx, paidOrder); err != nil {
		logger.Error(ctx, "Failed to refund payment after order update failure; manual refund required",
			zap.Error(err),
			zap.String("order_uuid", paidOrder.OrderUUID.String()),
			zap.String("transaction_uuid", paidOrder.TransactionUUID.String()),
		)
		return
	}

	logger.Warn(ctx, "Payment refunded after order update failure",
		zap.String("order_uuid", paidOrder.OrderUUID.String()),
		zap.String("transaction_uuid", paidOrder.TransactionUUID.String()),
	)
}
