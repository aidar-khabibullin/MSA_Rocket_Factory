package app

import (
	"context"
	"time"

	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"

	"github.com/aidar-khabibullin/MSA_Rocket_Factory/order/internal/config"
	"github.com/aidar-khabibullin/MSA_Rocket_Factory/platform/pkg/closer"
	wrappedKafka "github.com/aidar-khabibullin/MSA_Rocket_Factory/platform/pkg/kafka"
	wrappedKafkaConsumer "github.com/aidar-khabibullin/MSA_Rocket_Factory/platform/pkg/kafka/consumer"
	"github.com/aidar-khabibullin/MSA_Rocket_Factory/platform/pkg/logger"
	kafkaMiddleware "github.com/aidar-khabibullin/MSA_Rocket_Factory/platform/pkg/middleware/kafka"
	eventsV1 "github.com/aidar-khabibullin/MSA_Rocket_Factory/shared/pkg/proto/events/v1"
)

// dedupMiddleware возвращает middleware дедупликации ship.assembled по event_uuid.
// Если REDIS_DEDUP_ADDR не задан, дедупликация отключена и возвращается nil.
func (d *diContainer) dedupMiddleware(ctx context.Context, metrics *kafkaMiddleware.KafkaMetrics) wrappedKafkaConsumer.Middleware {
	cfg := config.AppConfig().Dedup
	if cfg.Addr() == "" {
		logger.Warn(ctx, "REDIS_DEDUP_ADDR не задан: дедупликация событий отключена")
		return nil
	}

	store, pool := wrappedKafkaConsumer.NewRedisDedupStoreFromAddr(cfg.Addr(), 2*time.Second)
	closer.AddNamed("Dedup Redis pool", func(context.Context) error { return pool.Close() })

	hooks := kafkaMiddleware.DedupMetricsHooks(metrics, "order", config.AppConfig().ShipAssembledConsumer.GroupID())

	logger.Info(ctx, "Дедупликация событий включена", zap.String("redis", cfg.Addr()), zap.Duration("ttl", cfg.TTL()))
	return wrappedKafkaConsumer.Dedup(store, "order", shipAssembledEventID, cfg.TTL(), logger.Logger(), hooks)
}

func shipAssembledEventID(msg wrappedKafka.Message) (string, error) {
	var event eventsV1.ShipAssembled
	if err := proto.Unmarshal(msg.Value, &event); err != nil {
		return "", err
	}
	return event.GetEventUuid(), nil
}
