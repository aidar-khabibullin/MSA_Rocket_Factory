package consumer

import (
	"context"
	"strconv"
	"time"

	"github.com/IBM/sarama"

	"github.com/aidar-khabibullin/MSA_Rocket_Factory/platform/pkg/kafka"
)

// DLQSuffix — суффикс топика с необработанными сообщениями: <исходный топик>.dlq.
const DLQSuffix = ".dlq"

// DeadLetterPublisher отправляет сообщение, которое не удалось обработать, в dead letter queue.
type DeadLetterPublisher interface {
	Publish(ctx context.Context, msg kafka.Message, cause error) error
}

// Resilience задаёт политику обработки ошибок consumer'а.
//
// Без DLQ действует прежнее поведение: после ошибки сообщение пропускается без отметки offset.
// С DLQ сообщение повторяется MaxRetries раз с экспоненциальной задержкой, затем уходит в DLQ,
// и только после этого offset сдвигается.
type Resilience struct {
	// MaxRetries — число повторов после первой неудачной попытки.
	MaxRetries int
	// InitialBackoff — задержка перед первым повтором, дальше удваивается до MaxBackoff.
	InitialBackoff time.Duration
	MaxBackoff     time.Duration
	DLQ            DeadLetterPublisher

	// Необязательные хуки для метрик.
	OnRetry      func(ctx context.Context, msg kafka.Message, attempt int, err error)
	OnDeadLetter func(ctx context.Context, msg kafka.Message, err error)
}

// DefaultResilience: 3 повтора с задержками 1 с, 2 с, 4 с.
func DefaultResilience(dlq DeadLetterPublisher) Resilience {
	return Resilience{MaxRetries: 3, InitialBackoff: time.Second, MaxBackoff: 4 * time.Second, DLQ: dlq}
}

// backoff возвращает задержку перед повтором с номером attempt (с нуля): initial * 2^attempt, не больше MaxBackoff.
func (r Resilience) backoff(attempt int) time.Duration {
	d := r.InitialBackoff
	for i := 0; i < attempt; i++ {
		d *= 2
		if r.MaxBackoff > 0 && d >= r.MaxBackoff {
			return r.MaxBackoff
		}
	}
	if r.MaxBackoff > 0 && d > r.MaxBackoff {
		return r.MaxBackoff
	}
	return d
}

// sleep ждёт d или отмены контекста.
func sleep(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return ctx.Err()
	}
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

type dlqPublisher struct {
	producer sarama.SyncProducer
}

// NewDLQPublisher создаёт публикатор, который пишет в топик "<исходный>.dlq".
// В заголовки добавляются исходные topic/partition/offset и текст ошибки.
func NewDLQPublisher(producer sarama.SyncProducer) DeadLetterPublisher {
	return &dlqPublisher{producer: producer}
}

func (p *dlqPublisher) Publish(_ context.Context, msg kafka.Message, cause error) error {
	headers := []sarama.RecordHeader{
		{Key: []byte("x-original-topic"), Value: []byte(msg.Topic)},
		{Key: []byte("x-original-partition"), Value: []byte(strconv.Itoa(int(msg.Partition)))},
		{Key: []byte("x-original-offset"), Value: []byte(strconv.FormatInt(msg.Offset, 10))},
		{Key: []byte("x-error"), Value: []byte(cause.Error())},
		{Key: []byte("x-failed-at"), Value: []byte(time.Now().UTC().Format(time.RFC3339))},
	}
	for k, v := range msg.Headers {
		headers = append(headers, sarama.RecordHeader{Key: []byte(k), Value: v})
	}

	_, _, err := p.producer.SendMessage(&sarama.ProducerMessage{
		Topic:   msg.Topic + DLQSuffix,
		Key:     sarama.ByteEncoder(msg.Key),
		Value:   sarama.ByteEncoder(msg.Value),
		Headers: headers,
	})
	return err
}
