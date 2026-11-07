package env

import (
	"time"

	"github.com/caarlos0/env/v11"
)

type dedupEnvConfig struct {
	// Addr — адрес Redis (host:port). Пусто — дедупликация отключена.
	Addr string        `env:"REDIS_DEDUP_ADDR"`
	TTL  time.Duration `env:"DEDUP_TTL" envDefault:"24h"`
}

type dedupConfig struct {
	raw dedupEnvConfig
}

func NewDedupConfig() (*dedupConfig, error) {
	var raw dedupEnvConfig
	if err := env.Parse(&raw); err != nil {
		return nil, err
	}

	return &dedupConfig{raw: raw}, nil
}

func (cfg *dedupConfig) Addr() string { return cfg.raw.Addr }

func (cfg *dedupConfig) TTL() time.Duration { return cfg.raw.TTL }
