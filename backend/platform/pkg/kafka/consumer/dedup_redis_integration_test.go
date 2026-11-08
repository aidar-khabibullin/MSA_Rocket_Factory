//go:build integration

package consumer

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/aidar-khabibullin/MSA_Rocket_Factory/platform/pkg/kafka"
)

func startRedis(t *testing.T) string {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	t.Cleanup(cancel)

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "redis:7.2.5-alpine3.20",
			ExposedPorts: []string{"6379/tcp"},
			WaitingFor:   wait.ForLog("Ready to accept connections").WithStartupTimeout(60 * time.Second),
		},
		Started: true,
	})
	require.NoError(t, err)
	t.Cleanup(func() { _ = container.Terminate(context.Background()) })

	endpoint, err := container.Endpoint(ctx, "")
	require.NoError(t, err)
	return endpoint
}

func TestRedisDedupStore(t *testing.T) {
	store, closer := NewRedisDedupStoreFromAddr(startRedis(t), 2*time.Second)
	t.Cleanup(func() { _ = closer.Close() })
	ctx := context.Background()

	seen, err := store.Seen(ctx, "dedup:x:1")
	require.NoError(t, err)
	require.False(t, seen)

	require.NoError(t, store.Mark(ctx, "dedup:x:1", time.Hour))
	seen, err = store.Seen(ctx, "dedup:x:1")
	require.NoError(t, err)
	require.True(t, seen)

	seen, err = store.Seen(ctx, "dedup:x:2")
	require.NoError(t, err)
	require.False(t, seen, "ключи разных событий не пересекаются")
}

func TestRedisDedupStore_KeysExpire(t *testing.T) {
	store, closer := NewRedisDedupStoreFromAddr(startRedis(t), 2*time.Second)
	t.Cleanup(func() { _ = closer.Close() })
	ctx := context.Background()

	require.NoError(t, store.Mark(ctx, "dedup:x:ttl", 300*time.Millisecond))

	require.Eventually(t, func() bool {
		seen, err := store.Seen(ctx, "dedup:x:ttl")
		return err == nil && !seen
	}, 3*time.Second, 100*time.Millisecond, "ключ должен исчезнуть по TTL, иначе Redis копит ключи вечно")
}

// Сквозной сценарий с настоящим Redis: повторная доставка одного event_uuid не вызывает обработчик второй раз.
func TestDedupMiddleware_WithRealRedis(t *testing.T) {
	store, closer := NewRedisDedupStoreFromAddr(startRedis(t), 2*time.Second)
	t.Cleanup(func() { _ = closer.Close() })

	calls := 0
	handler := Dedup(store, "assembly", func(msg kafka.Message) (string, error) { return string(msg.Value), nil },
		time.Hour, nopLogger{}, DedupHooks{})(func(context.Context, kafka.Message) error { calls++; return nil })

	msg := kafka.Message{Value: []byte("event-1")}
	for range 3 {
		require.NoError(t, handler(context.Background(), msg))
	}

	require.Equal(t, 1, calls)
}
