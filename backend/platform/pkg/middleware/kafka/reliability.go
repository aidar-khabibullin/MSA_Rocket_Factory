package kafka

import (
	"context"

	wrapped "github.com/aidar-khabibullin/MSA_Rocket_Factory/platform/pkg/kafka"
	"github.com/aidar-khabibullin/MSA_Rocket_Factory/platform/pkg/kafka/consumer"
)

// ConsumerResilience возвращает политику повторов и DLQ по умолчанию с хуками метрик.
// metrics может быть nil: тогда метрики просто не пишутся.
func ConsumerResilience(metrics *KafkaMetrics, groupID string, dlq consumer.DeadLetterPublisher) consumer.Resilience {
	resilience := consumer.DefaultResilience(dlq)
	if metrics == nil {
		return resilience
	}

	resilience.OnRetry = func(ctx context.Context, msg wrapped.Message, _ int, _ error) {
		metrics.RecordConsumerRetry(ctx, msg.Topic, groupID)
	}
	resilience.OnDeadLetter = func(ctx context.Context, msg wrapped.Message, _ error) {
		metrics.RecordConsumerDeadLetter(ctx, msg.Topic, groupID)
	}
	return resilience
}

// DedupMetricsHooks возвращает хуки дедупликации, пишущие метрики. metrics может быть nil.
func DedupMetricsHooks(metrics *KafkaMetrics, scope, groupID string) consumer.DedupHooks {
	if metrics == nil {
		return consumer.DedupHooks{}
	}

	return consumer.DedupHooks{
		OnDuplicate: func(ctx context.Context, msg wrapped.Message) {
			metrics.RecordDedupDuplicate(ctx, msg.Topic, groupID)
		},
		OnStoreError: func(ctx context.Context, _ error) {
			metrics.RecordDedupStoreError(ctx, scope)
		},
	}
}
