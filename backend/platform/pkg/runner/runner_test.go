package runner_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aidar-khabibullin/MSA_Rocket_Factory/platform/pkg/runner"
)

// waitDone ждёт отмены контекста, как это делает настоящая долгоживущая задача
func waitDone(ctx context.Context) error {
	<-ctx.Done()
	return ctx.Err()
}

func TestRun_NoTasks(t *testing.T) {
	require.NoError(t, runner.Run(context.Background()))
}

func TestRun_ParentCancelStopsAllWithoutError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error, 1)
	go func() {
		done <- runner.Run(ctx,
			runner.Task{Name: "a", Run: waitDone},
			runner.Task{Name: "b", Run: waitDone},
		)
	}()

	cancel()

	select {
	case err := <-done:
		assert.NoError(t, err, "штатная остановка по сигналу не должна быть ошибкой")
	case <-time.After(2 * time.Second):
		t.Fatal("Run не завершился после отмены контекста")
	}
}

// Регрессия: раньше ошибка консьюмера только логировалась, а сервис продолжал жить без него
func TestRun_FailedTaskStopsOthersAndReturnsError(t *testing.T) {
	boom := errors.New("kafka: client has run out of available brokers")
	otherStopped := make(chan struct{})

	done := make(chan error, 1)
	go func() {
		done <- runner.Run(context.Background(),
			runner.Task{Name: "consumer", Run: func(context.Context) error { return boom }},
			runner.Task{Name: "http", Run: func(ctx context.Context) error {
				err := waitDone(ctx)
				close(otherStopped)
				return err
			}},
		)
	}()

	select {
	case err := <-done:
		require.Error(t, err)
		assert.ErrorIs(t, err, boom)
		assert.Contains(t, err.Error(), "consumer", "в ошибке должно быть имя упавшей задачи")
	case <-time.After(2 * time.Second):
		t.Fatal("Run не завершился после падения задачи")
	}

	select {
	case <-otherStopped:
	case <-time.After(2 * time.Second):
		t.Fatal("вторая задача не получила отмену контекста")
	}
}

func TestRun_CleanExitOfOneTaskStopsOthersWithoutError(t *testing.T) {
	err := runner.Run(context.Background(),
		runner.Task{Name: "finishes", Run: func(context.Context) error { return nil }},
		runner.Task{Name: "waits", Run: waitDone},
	)

	assert.NoError(t, err)
}

func TestRun_FirstErrorWins(t *testing.T) {
	first := errors.New("first")
	second := errors.New("second")
	firstFailed := make(chan struct{})

	err := runner.Run(context.Background(),
		runner.Task{Name: "a", Run: func(context.Context) error {
			defer close(firstFailed)
			return first
		}},
		runner.Task{Name: "b", Run: func(context.Context) error {
			<-firstFailed
			return second
		}},
	)

	require.Error(t, err)
	assert.ErrorIs(t, err, first)
	assert.NotErrorIs(t, err, second)
}

type fakeConsumer struct {
	run func(ctx context.Context) error
}

func (f fakeConsumer) RunConsumer(ctx context.Context) error { return f.run(ctx) }

func TestConsumerTask_ErrorIsPropagated(t *testing.T) {
	boom := errors.New("kafka down")

	err := runner.ConsumerTask("orders", fakeConsumer{run: func(context.Context) error { return boom }}).
		Run(context.Background())

	assert.ErrorIs(t, err, boom)
}

// Консьюмер вернул nil при живом контексте: он молча перестал работать - это ошибка
func TestConsumerTask_UnexpectedStopIsError(t *testing.T) {
	err := runner.ConsumerTask("orders", fakeConsumer{run: func(context.Context) error { return nil }}).
		Run(context.Background())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "stopped unexpectedly")
}

func TestConsumerTask_StopOnCancelIsClean(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := runner.ConsumerTask("orders", fakeConsumer{run: func(ctx context.Context) error {
		<-ctx.Done()
		return nil // sarama возвращает nil при закрытии группы
	}}).Run(ctx)

	assert.NoError(t, err)
}
