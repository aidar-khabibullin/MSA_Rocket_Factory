//go:build integration

package integration

import (
	"context"

	"go.uber.org/zap"

	"github.com/aidar-khabibullin/MSA_Rocket_Factory/platform/pkg/logger"
)

// teardownTestEnvironment — освобождает все ресурсы тестового окружения
func teardownTestEnvironment(ctx context.Context, env *TestEnvironment) {
	log := logger.Logger()
	log.Info(ctx, "🧹 Очистка тестового окружения...")

	cleanupTestEnvironment(ctx, env)

	log.Info(ctx, "✅ Тестовое окружение успешно очищено")
}

// cleanupTestEnvironment — вспомогательная функция для освобождения ресурсов
func cleanupTestEnvironment(ctx context.Context, env *TestEnvironment) {
	if env.DBPool != nil {
		env.DBPool.Close()
		logger.Info(ctx, "🛑 Пул подключений к PostgreSQL закрыт")
	}

	if env.App != nil {
		if err := env.App.Terminate(ctx); err != nil {
			logger.Error(ctx, "не удалось остановить контейнер приложения", zap.Error(err))
		} else {
			logger.Info(ctx, "🛑 Контейнер приложения остановлен")
		}
	}

	if env.Postgres != nil {
		if err := env.Postgres.Terminate(ctx); err != nil {
			logger.Error(ctx, "не удалось остановить контейнер PostgreSQL", zap.Error(err))
		} else {
			logger.Info(ctx, "🛑 Контейнер PostgreSQL остановлен")
		}
	}

	if env.Network != nil {
		if err := env.Network.Remove(ctx); err != nil {
			logger.Error(ctx, "не удалось удалить сеть", zap.Error(err))
		} else {
			logger.Info(ctx, "🛑 Сеть удалена")
		}
	}
}
