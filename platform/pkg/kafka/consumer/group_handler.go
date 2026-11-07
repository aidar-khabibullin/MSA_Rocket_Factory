package consumer

import (
	"context"
	"fmt"

	"github.com/IBM/sarama"
	"go.uber.org/zap"

	"github.com/aidar-khabibullin/MSA_Rocket_Factory/platform/pkg/kafka"
)

// Middleware — функция middleware для дополнительной обработки.
type Middleware func(next kafka.MessageHandler) kafka.MessageHandler

// groupHandler — обёртка для sarama.ConsumerGroupHandler
type groupHandler struct {
	handler    kafka.MessageHandler
	logger     Logger
	resilience Resilience
}

// Реализуем интерфейс kafka.ConsumerGroupHandler
var _ kafka.ConsumerGroupHandler = (*groupHandler)(nil)

// NewGroupHandler создаёт новый groupHandler с middleware цепочкой.
func NewGroupHandler(handler kafka.MessageHandler, logger Logger, middlewares ...Middleware) *groupHandler {
	return NewGroupHandlerWithResilience(handler, logger, Resilience{}, middlewares...)
}

// NewGroupHandlerWithResilience создаёт groupHandler с политикой повторов и DLQ.
func NewGroupHandlerWithResilience(handler kafka.MessageHandler, logger Logger, resilience Resilience, middlewares ...Middleware) *groupHandler {
	// Применяем middleware цепочку
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}

	return &groupHandler{
		handler:    handler,
		logger:     logger,
		resilience: resilience,
	}
}

// NewGroupHandlerWithConsumerGroupMiddleware создаёт новый groupHandler с middleware цепочкой и consumer group middleware.
func NewGroupHandlerWithConsumerGroupMiddleware(handler kafka.MessageHandler, logger Logger, messageMiddlewares []Middleware, consumerGroupMiddlewares []kafka.ConsumerGroupHandlerMiddleware) kafka.ConsumerGroupHandler {
	return newGroupHandlerWithConsumerGroupMiddleware(handler, logger, Resilience{}, messageMiddlewares, consumerGroupMiddlewares)
}

func newGroupHandlerWithConsumerGroupMiddleware(handler kafka.MessageHandler, logger Logger, resilience Resilience, messageMiddlewares []Middleware, consumerGroupMiddlewares []kafka.ConsumerGroupHandlerMiddleware) kafka.ConsumerGroupHandler {
	// Применяем middleware цепочку для сообщений
	for i := len(messageMiddlewares) - 1; i >= 0; i-- {
		handler = messageMiddlewares[i](handler)
	}

	// Создаем базовый handler
	baseHandler := &groupHandler{
		handler:    handler,
		logger:     logger,
		resilience: resilience,
	}

	// Применяем consumer group middleware цепочку
	result := kafka.ConsumerGroupHandler(baseHandler)
	for i := len(consumerGroupMiddlewares) - 1; i >= 0; i-- {
		result = consumerGroupMiddlewares[i](result)
	}

	return result
}

func (g *groupHandler) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

func (g *groupHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (g *groupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message, ok := <-claim.Messages():
			if !ok {
				g.logger.Info(session.Context(), "Kafka message channel closed")
				return nil
			}

			msg := kafka.Message{
				Key:            message.Key,
				Value:          message.Value,
				Topic:          message.Topic,
				Partition:      message.Partition,
				Offset:         message.Offset,
				Timestamp:      message.Timestamp,
				BlockTimestamp: message.BlockTimestamp,
				Headers:        extractHeaders(message.Headers),
			}

			marked, err := g.process(session.Context(), msg)
			if err != nil {
				// Не удалось ни обработать, ни отправить в DLQ (или сервис останавливается):
				// завершаем claim без отметки offset, сообщение будет доставлено заново.
				return err
			}
			if !marked {
				continue
			}

			session.MarkMessage(message, "")

		case <-session.Context().Done():
			g.logger.Info(session.Context(), "Kafka session context done")
			return nil
		}
	}
}

func extractHeaders(headers []*sarama.RecordHeader) map[string][]byte {
	result := make(map[string][]byte)
	for _, h := range headers {
		if h != nil && h.Key != nil {
			result[string(h.Key)] = h.Value
		}
	}

	return result
}

// process обрабатывает сообщение с повторами. Возвращает marked=true, если offset можно сдвигать
// (обработано успешно или отправлено в DLQ). Ошибка означает, что claim нужно прервать.
func (g *groupHandler) process(ctx context.Context, msg kafka.Message) (marked bool, err error) {
	var handlerErr error
	for attempt := 0; ; attempt++ {
		handlerErr = g.handler(ctx, msg)
		if handlerErr == nil {
			return true, nil
		}
		g.logger.Error(ctx, "Kafka handler error", zap.Error(handlerErr),
			zap.String("topic", msg.Topic), zap.Int32("partition", msg.Partition),
			zap.Int64("offset", msg.Offset), zap.Int("attempt", attempt+1))

		if g.resilience.DLQ == nil || attempt >= g.resilience.MaxRetries {
			break
		}
		if g.resilience.OnRetry != nil {
			g.resilience.OnRetry(ctx, msg, attempt+1, handlerErr)
		}
		if err := sleep(ctx, g.resilience.backoff(attempt)); err != nil {
			// Сервис останавливается: сообщение не обработано и не должно попасть в DLQ.
			return false, err
		}
	}

	if g.resilience.DLQ == nil {
		// Прежнее поведение без DLQ: сообщение пропускается без отметки offset.
		return false, nil
	}

	if err := g.resilience.DLQ.Publish(ctx, msg, handlerErr); err != nil {
		g.logger.Error(ctx, "Failed to publish message to DLQ", zap.Error(err),
			zap.String("topic", msg.Topic), zap.Int64("offset", msg.Offset))
		return false, fmt.Errorf("publish to DLQ: %w", err)
	}
	if g.resilience.OnDeadLetter != nil {
		g.resilience.OnDeadLetter(ctx, msg, handlerErr)
	}
	g.logger.Error(ctx, "Message sent to DLQ after retries exhausted", zap.Error(handlerErr),
		zap.String("topic", msg.Topic+DLQSuffix), zap.Int32("partition", msg.Partition), zap.Int64("offset", msg.Offset))

	return true, nil
}
