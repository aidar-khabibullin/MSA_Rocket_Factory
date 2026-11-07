package consumer

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/aidar-khabibullin/MSA_Rocket_Factory/platform/pkg/kafka"
)

type memStore struct {
	seen    map[string]bool
	seenErr error
	markErr error
	ttl     time.Duration
}

func newMemStore() *memStore { return &memStore{seen: map[string]bool{}} }

func (s *memStore) Seen(_ context.Context, key string) (bool, error) {
	return s.seen[key], s.seenErr
}

func (s *memStore) Mark(_ context.Context, key string, ttl time.Duration) error {
	if s.markErr != nil {
		return s.markErr
	}
	s.seen[key] = true
	s.ttl = ttl
	return nil
}

func eventID(msg kafka.Message) (string, error) { return string(msg.Value), nil }

func dedup(store DedupStore, next kafka.MessageHandler, hooks DedupHooks) kafka.MessageHandler {
	return Dedup(store, "assembly", eventID, time.Hour, nopLogger{}, hooks)(next)
}

func TestDedup_SecondDeliveryIsSkipped(t *testing.T) {
	calls := 0
	store := newMemStore()
	h := dedup(store, func(context.Context, kafka.Message) error { calls++; return nil }, DedupHooks{})

	msg := kafka.Message{Value: []byte("evt-1")}
	require.NoError(t, h(context.Background(), msg))
	require.NoError(t, h(context.Background(), msg))

	require.Equal(t, 1, calls, "повторная доставка с тем же event_uuid не должна обрабатываться")
	require.Equal(t, time.Hour, store.ttl)
}

func TestDedup_FailedHandlerIsNotMarked(t *testing.T) {
	calls := 0
	store := newMemStore()
	h := dedup(store, func(context.Context, kafka.Message) error {
		calls++
		if calls == 1 {
			return errors.New("boom")
		}
		return nil
	}, DedupHooks{})

	msg := kafka.Message{Value: []byte("evt-1")}
	require.Error(t, h(context.Background(), msg))
	require.NoError(t, h(context.Background(), msg), "повтор после ошибки должен обрабатываться")
	require.Equal(t, 2, calls)
}

func TestDedup_DifferentEventsAreBothProcessed(t *testing.T) {
	calls := 0
	h := dedup(newMemStore(), func(context.Context, kafka.Message) error { calls++; return nil }, DedupHooks{})

	require.NoError(t, h(context.Background(), kafka.Message{Value: []byte("evt-1")}))
	require.NoError(t, h(context.Background(), kafka.Message{Value: []byte("evt-2")}))
	require.Equal(t, 2, calls)
}

// Redis недоступен: обработка продолжается (fail-open), дубликат в этот момент не отсекается.
func TestDedup_StoreUnavailable_FailsOpen(t *testing.T) {
	calls := 0
	store := newMemStore()
	store.seenErr = errors.New("redis down")
	store.markErr = errors.New("redis down")
	var storeErrors int
	h := dedup(store, func(context.Context, kafka.Message) error { calls++; return nil },
		DedupHooks{OnStoreError: func(context.Context, error) { storeErrors++ }})

	msg := kafka.Message{Value: []byte("evt-1")}
	require.NoError(t, h(context.Background(), msg))
	require.NoError(t, h(context.Background(), msg))

	require.Equal(t, 2, calls)
	require.Positive(t, storeErrors, "недоступность хранилища должна попадать в метрики")
}

func TestDedup_UnreadableEventIDPassesThrough(t *testing.T) {
	calls := 0
	h := Dedup(newMemStore(), "x", func(kafka.Message) (string, error) { return "", errors.New("bad payload") },
		time.Hour, nopLogger{}, DedupHooks{})(func(context.Context, kafka.Message) error { calls++; return nil })

	require.NoError(t, h(context.Background(), kafka.Message{}))
	require.Equal(t, 1, calls, "решение о битом сообщении принимает сам обработчик")
}

func TestDedup_DuplicateHook(t *testing.T) {
	dups := 0
	h := dedup(newMemStore(), func(context.Context, kafka.Message) error { return nil },
		DedupHooks{OnDuplicate: func(context.Context, kafka.Message) { dups++ }})

	msg := kafka.Message{Value: []byte("evt-1")}
	_ = h(context.Background(), msg)
	_ = h(context.Background(), msg)

	require.Equal(t, 1, dups)
}
