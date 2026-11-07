package consumer

import (
	"context"
	"time"

	"github.com/gomodule/redigo/redis"
	"go.uber.org/zap"

	"github.com/aidar-khabibullin/MSA_Rocket_Factory/platform/pkg/kafka"
)

// DedupStore хранит идентификаторы уже обработанных событий.
type DedupStore interface {
	Seen(ctx context.Context, key string) (bool, error)
	Mark(ctx context.Context, key string, ttl time.Duration) error
}

// EventIDFunc достаёт идентификатор события (event_uuid) из сообщения.
type EventIDFunc func(msg kafka.Message) (string, error)

// DedupHooks — необязательные хуки для метрик.
type DedupHooks struct {
	OnDuplicate  func(ctx context.Context, msg kafka.Message)
	OnStoreError func(ctx context.Context, err error)
}

// Dedup возвращает middleware, пропускающий повторную доставку уже обработанного события.
//
// Порядок такой: проверить -> обработать -> отметить. Отметка ставится только после успешной
// обработки, поэтому упавшее сообщение будет повторено. Остаётся окно «обработано, но процесс упал
// до отметки»: событие тогда обработается второй раз (гарантия at-least-once, не exactly-once).
//
// Если хранилище недоступно, обработка продолжается (fail-open): потеря сообщения хуже дубликата,
// а получатели событий должны быть идемпотентны. Ошибка попадает в лог и хук OnStoreError.
func Dedup(store DedupStore, scope string, eventID EventIDFunc, ttl time.Duration, logger Logger, hooks DedupHooks) Middleware {
	return func(next kafka.MessageHandler) kafka.MessageHandler {
		return func(ctx context.Context, msg kafka.Message) error {
			id, err := eventID(msg)
			if err != nil || id == "" {
				// Не можем определить событие: пусть обработчик сам решит, что делать с таким сообщением.
				return next(ctx, msg)
			}
			key := "dedup:" + scope + ":" + id

			seen, err := store.Seen(ctx, key)
			if err != nil {
				reportStoreError(ctx, logger, hooks, "check", err)
			} else if seen {
				logger.Info(ctx, "Duplicate event skipped", zap.String("event_uuid", id), zap.String("scope", scope))
				if hooks.OnDuplicate != nil {
					hooks.OnDuplicate(ctx, msg)
				}
				return nil
			}

			if err := next(ctx, msg); err != nil {
				return err
			}

			if err := store.Mark(ctx, key, ttl); err != nil {
				reportStoreError(ctx, logger, hooks, "mark", err)
			}
			return nil
		}
	}
}

func reportStoreError(ctx context.Context, logger Logger, hooks DedupHooks, op string, err error) {
	logger.Error(ctx, "Dedup store unavailable, processing without deduplication", zap.String("op", op), zap.Error(err))
	if hooks.OnStoreError != nil {
		hooks.OnStoreError(ctx, err)
	}
}

type redisDedupStore struct {
	pool *redis.Pool
}

// NewRedisDedupStore создаёт хранилище дедупликации поверх пула redigo.
func NewRedisDedupStore(pool *redis.Pool) DedupStore {
	return &redisDedupStore{pool: pool}
}

func (s *redisDedupStore) Seen(ctx context.Context, key string) (bool, error) {
	conn, err := s.pool.GetContext(ctx)
	if err != nil {
		return false, err
	}
	defer conn.Close()

	return redis.Bool(conn.Do("EXISTS", key))
}

func (s *redisDedupStore) Mark(ctx context.Context, key string, ttl time.Duration) error {
	conn, err := s.pool.GetContext(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.Do("SET", key, "1", "PX", ttl.Milliseconds())
	return err
}

// NewRedisDedupStoreFromAddr создаёт хранилище дедупликации, подключаясь к Redis по адресу host:port.
// Соединения устанавливаются лениво, поэтому недоступный Redis не мешает запуску сервиса.
// Второй результат нужно закрыть при остановке сервиса.
func NewRedisDedupStoreFromAddr(addr string, dialTimeout time.Duration) (DedupStore, interface{ Close() error }) {
	pool := &redis.Pool{
		MaxIdle:     4,
		MaxActive:   16,
		IdleTimeout: 5 * time.Minute,
		Wait:        true,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", addr,
				redis.DialConnectTimeout(dialTimeout),
				redis.DialReadTimeout(dialTimeout),
				redis.DialWriteTimeout(dialTimeout),
			)
		},
	}
	return NewRedisDedupStore(pool), pool
}
