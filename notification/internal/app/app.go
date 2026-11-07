package app

import (
	"context"

	"go.uber.org/zap"

	"github.com/aidar-khabibullin/MSA_Rocket_Factory/notification/internal/config"
	"github.com/aidar-khabibullin/MSA_Rocket_Factory/platform/pkg/closer"
	"github.com/aidar-khabibullin/MSA_Rocket_Factory/platform/pkg/logger"
	"github.com/aidar-khabibullin/MSA_Rocket_Factory/platform/pkg/runner"
	"github.com/aidar-khabibullin/MSA_Rocket_Factory/platform/pkg/tracing"
)

type App struct {
	diContainer *diContainer
}

func New(ctx context.Context) (*App, error) {
	a := &App{}

	err := a.initDeps(ctx)
	if err != nil {
		return nil, err
	}

	return a, nil
}

func (a *App) Run(ctx context.Context) error {
	// Kafka-консьюмеры критичны: без них сервис не выполняет свою работу. Если любой из них упал,
	// падает весь сервис, и Docker (restart: unless-stopped) его перезапускает.
	// Раньше ошибка только логировалась, и сервис оставался живым процессом без консьюмеров.
	err := runner.Run(ctx,
		runner.ConsumerTask("OrderPaid Consumer", a.diContainer.OrderPaidConsumer(ctx)),
		runner.ConsumerTask("ShipAssembled Consumer", a.diContainer.ShipAssembledConsumer(ctx)),
		// Telegram-бот нефатален: в режиме разработки используется заглушка, и его сбой не должен ронять сервис
		runner.Task{Name: "Telegram bot", Run: func(ctx context.Context) error {
			if err := a.diContainer.TelegramClient(ctx).Start(ctx); err != nil {
				logger.Error(ctx, "❌ Ошибка при запуске Telegram бота", zap.Error(err))
			}

			<-ctx.Done()

			return nil
		}},
	)

	if ctx.Err() != nil {
		logger.Info(ctx, "🛑 Получен сигнал завершения работы")
	}

	return err
}

func (a *App) initDeps(ctx context.Context) error {
	inits := []func(context.Context) error{
		a.initDI,
		a.initLogger,
		a.initTracing,
		a.initCloser,
	}

	for _, f := range inits {
		err := f(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *App) initDI(_ context.Context) error {
	a.diContainer = NewDiContainer()
	return nil
}

func (a *App) initLogger(ctx context.Context) error {
	return logger.Init(
		ctx,
		config.AppConfig().Logger.Level(),
		config.AppConfig().Logger.AsJson(),
		config.AppConfig().Logger.Outputs(),
		config.AppConfig().Logger.OtelEndpoint(),
		config.AppConfig().Logger.ServiceName(),
		"dev", // serviceEnvironment - хардкод для обратной совместимости
	)
}

func (a *App) initTracing(ctx context.Context) error {
	return tracing.InitTracer(ctx, config.AppConfig().Tracing)
}

func (a *App) initCloser(_ context.Context) error {
	closer.SetLogger(logger.Logger())

	// Добавляем graceful shutdown для tracing
	closer.AddNamed("Tracing provider", tracing.ShutdownTracer)

	return nil
}

// DiContainer возвращает DI контейнер для тестирования
func (a *App) DiContainer() *diContainer {
	return a.diContainer
}
