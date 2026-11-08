package config

import (
	"time"

	"github.com/IBM/sarama"
)

type LoggerConfig interface {
	Level() string
	AsJson() bool
	Outputs() string
	OtelEndpoint() string
	ServiceName() string
}

type KafkaConfig interface {
	Brokers() []string
}

type OrderPaidConsumerConfig interface {
	Topic() string
	GroupID() string
	Config() *sarama.Config
}

type ShipAssemblyConsumerConfig interface {
	Topic() string
	GroupID() string
	Config() *sarama.Config
}

type TelegramConfig interface {
	BotToken() string
	SkipAPICheck() bool
}

type GRPCClientConfig interface {
	IAMAddress() string
}

type TracingConfig interface {
	CollectorEndpoint() string
	ServiceName() string
	Environment() string
	ServiceVersion() string
}

// DedupConfig — настройки дедупликации Kafka-событий.
type DedupConfig interface {
	// Addr — адрес Redis; пусто — дедупликация отключена.
	Addr() string
	TTL() time.Duration
}
