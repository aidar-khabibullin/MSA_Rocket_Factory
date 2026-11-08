package consumer

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/IBM/sarama"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/aidar-khabibullin/MSA_Rocket_Factory/platform/pkg/kafka"
)

type nopLogger struct{}

func (nopLogger) Info(context.Context, string, ...zap.Field)  {}
func (nopLogger) Error(context.Context, string, ...zap.Field) {}

type fakeSession struct {
	sarama.ConsumerGroupSession
	ctx    context.Context //nolint:containedctx // тестовая заглушка sarama.ConsumerGroupSession должна отдавать контекст из Context()
	mu     sync.Mutex
	marked []int64
}

func (s *fakeSession) Context() context.Context { return s.ctx }
func (s *fakeSession) MarkMessage(m *sarama.ConsumerMessage, _ string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.marked = append(s.marked, m.Offset)
}

type fakeClaim struct {
	sarama.ConsumerGroupClaim
	ch chan *sarama.ConsumerMessage
}

func (c *fakeClaim) Messages() <-chan *sarama.ConsumerMessage { return c.ch }

type fakeDLQ struct {
	err  error
	got  []kafka.Message
	errs []error
}

func (d *fakeDLQ) Publish(_ context.Context, msg kafka.Message, cause error) error {
	d.got = append(d.got, msg)
	d.errs = append(d.errs, cause)
	return d.err
}

// run подаёт сообщения в claim, закрывает канал и возвращает результат ConsumeClaim.
func run(t *testing.T, h *groupHandler, offsets ...int64) (*fakeSession, error) {
	t.Helper()
	session := &fakeSession{ctx: context.Background()}
	claim := &fakeClaim{ch: make(chan *sarama.ConsumerMessage, len(offsets))}
	for _, o := range offsets {
		claim.ch <- &sarama.ConsumerMessage{Topic: "orders", Partition: 2, Offset: o, Key: []byte("k"), Value: []byte("v")}
	}
	close(claim.ch)
	return session, h.ConsumeClaim(session, claim)
}

func fastResilience(dlq DeadLetterPublisher, retries int) Resilience {
	return Resilience{MaxRetries: retries, InitialBackoff: time.Millisecond, MaxBackoff: 2 * time.Millisecond, DLQ: dlq}
}

func TestConsumeClaim_SuccessMarksOffset(t *testing.T) {
	h := NewGroupHandlerWithResilience(func(context.Context, kafka.Message) error { return nil }, nopLogger{}, fastResilience(&fakeDLQ{}, 3))

	session, err := run(t, h, 10, 11)

	require.NoError(t, err)
	require.Equal(t, []int64{10, 11}, session.marked)
}

func TestConsumeClaim_RetriesThenSucceeds(t *testing.T) {
	calls := 0
	dlq := &fakeDLQ{}
	h := NewGroupHandlerWithResilience(func(context.Context, kafka.Message) error {
		calls++
		if calls < 2 {
			return errors.New("temporary")
		}
		return nil
	}, nopLogger{}, fastResilience(dlq, 3))

	session, err := run(t, h, 10)

	require.NoError(t, err)
	require.Equal(t, 2, calls, "успех на второй попытке: третья не нужна")
	require.Empty(t, dlq.got, "сообщение не должно попасть в DLQ")
	require.Equal(t, []int64{10}, session.marked)
}

func TestConsumeClaim_ExhaustedRetriesGoToDLQAndConsumerMovesOn(t *testing.T) {
	boom := errors.New("permanent")
	attempts := map[int64]int{}
	dlq := &fakeDLQ{}
	h := NewGroupHandlerWithResilience(func(_ context.Context, m kafka.Message) error {
		attempts[m.Offset]++
		if m.Offset == 10 {
			return boom
		}
		return nil
	}, nopLogger{}, fastResilience(dlq, 3))

	session, err := run(t, h, 10, 11)

	require.NoError(t, err)
	require.Equal(t, 4, attempts[10], "1 попытка + 3 повтора")
	require.Len(t, dlq.got, 1)
	require.Equal(t, int64(10), dlq.got[0].Offset)
	require.ErrorIs(t, dlq.errs[0], boom)
	require.Equal(t, []int64{10, 11}, session.marked, "после DLQ offset сдвигается, следующее сообщение обработано")
}

func TestConsumeClaim_DLQFailureStopsClaimWithoutMarking(t *testing.T) {
	dlq := &fakeDLQ{err: errors.New("kafka down")}
	processed := 0
	h := NewGroupHandlerWithResilience(func(_ context.Context, m kafka.Message) error {
		processed++
		return errors.New("permanent")
	}, nopLogger{}, fastResilience(dlq, 1))

	session, err := run(t, h, 10, 11)

	require.Error(t, err, "claim прерывается, сообщения будут доставлены заново")
	require.Empty(t, session.marked, "offset не помечается, иначе сообщение будет потеряно")
	require.Equal(t, 2, processed, "второе сообщение не обрабатывалось")
}

func TestConsumeClaim_WithoutDLQKeepsLegacyBehaviour(t *testing.T) {
	h := NewGroupHandler(func(_ context.Context, m kafka.Message) error {
		if m.Offset == 10 {
			return errors.New("fail")
		}
		return nil
	}, nopLogger{})

	session, err := run(t, h, 10, 11)

	require.NoError(t, err)
	require.Equal(t, []int64{11}, session.marked)
}

func TestConsumeClaim_ContextCancelledDuringBackoffStopsWithoutDLQ(t *testing.T) {
	dlq := &fakeDLQ{}
	ctx, cancel := context.WithCancel(context.Background())
	h := NewGroupHandlerWithResilience(func(context.Context, kafka.Message) error {
		cancel()
		return errors.New("fail")
	}, nopLogger{}, Resilience{MaxRetries: 3, InitialBackoff: time.Hour, MaxBackoff: time.Hour, DLQ: dlq})

	session := &fakeSession{ctx: ctx}
	claim := &fakeClaim{ch: make(chan *sarama.ConsumerMessage, 1)}
	claim.ch <- &sarama.ConsumerMessage{Topic: "orders", Offset: 10}

	err := h.ConsumeClaim(session, claim)

	require.ErrorIs(t, err, context.Canceled)
	require.Empty(t, dlq.got, "при остановке сервиса сообщение не должно уходить в DLQ")
	require.Empty(t, session.marked)
}

func TestBackoff_DoublesAndCaps(t *testing.T) {
	r := Resilience{InitialBackoff: time.Second, MaxBackoff: 4 * time.Second}
	require.Equal(t, time.Second, r.backoff(0))
	require.Equal(t, 2*time.Second, r.backoff(1))
	require.Equal(t, 4*time.Second, r.backoff(2))
	require.Equal(t, 4*time.Second, r.backoff(5))
}
