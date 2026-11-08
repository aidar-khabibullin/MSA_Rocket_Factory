// Package runner запускает несколько долгоживущих задач сервиса (Kafka-консьюмеры, HTTP-сервер и т.п.)
// как единое целое: если одна из них упала, останавливаются все остальные, а ошибка возвращается наружу.
//
// Зачем: задача в горутине, которая только логирует ошибку, оставляет сервис "живым" процессом,
// который перестал выполнять свою работу (например, консьюмер потерял Kafka). Orchestrator
// (Docker restart policy, Kubernetes) такой сервис не перезапустит. Падение процесса делает проблему видимой
// и чинится перезапуском.
package runner

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

// Task - долгоживущая задача. Run должна работать, пока не отменён переданный ей контекст.
type Task struct {
	Name string
	Run  func(ctx context.Context) error
}

// Run запускает все задачи параллельно и ждёт завершения всех.
//
// Правила:
//   - задача вернула ошибку (кроме context.Canceled при отменённом контексте) - контекст остальных отменяется,
//     Run вернёт первую такую ошибку, обёрнутую именем задачи;
//   - задача вернула nil - это штатная остановка, остальные тоже получают отмену, ошибки нет;
//   - родительский контекст отменён - все задачи получают отмену, Run вернёт nil.
func Run(ctx context.Context, tasks ...Task) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var (
		wg       sync.WaitGroup
		mu       sync.Mutex
		firstErr error
	)

	for _, task := range tasks {
		wg.Add(1)

		go func() {
			defer wg.Done()

			err := task.Run(ctx)
			// Что бы ни случилось (ошибка или штатный выход), остальные задачи должны остановиться
			cancel()

			if err == nil || (errors.Is(err, context.Canceled) && ctx.Err() != nil) {
				return
			}

			mu.Lock()
			defer mu.Unlock()
			if firstErr == nil {
				firstErr = fmt.Errorf("%s: %w", task.Name, err)
			}
		}()
	}

	wg.Wait()

	return firstErr
}

// Consumer - то, что умеет читать сообщения до отмены контекста (Kafka-консьюмер).
type Consumer interface {
	RunConsumer(ctx context.Context) error
}

// ConsumerTask превращает консьюмер в Task. Неожиданная остановка консьюмера (он вернул nil,
// хотя контекст ещё не отменён) считается ошибкой: иначе сервис продолжил бы жить без консьюмера.
func ConsumerTask(name string, consumer Consumer) Task {
	return Task{
		Name: name,
		Run: func(ctx context.Context) error {
			if err := consumer.RunConsumer(ctx); err != nil {
				return err
			}

			if ctx.Err() == nil {
				return errors.New("consumer stopped unexpectedly")
			}

			return nil
		},
	}
}
