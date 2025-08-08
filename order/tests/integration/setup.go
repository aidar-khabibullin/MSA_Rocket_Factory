//go:build integration

package integration

import (
	"context"
	"os"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/zap"

	"github.com/aidar-khabibullin/MSA_Rocket_Factory/platform/pkg/logger"
	"github.com/aidar-khabibullin/MSA_Rocket_Factory/platform/pkg/testcontainers"
	"github.com/aidar-khabibullin/MSA_Rocket_Factory/platform/pkg/testcontainers/app"
	"github.com/aidar-khabibullin/MSA_Rocket_Factory/platform/pkg/testcontainers/network"
	"github.com/aidar-khabibullin/MSA_Rocket_Factory/platform/pkg/testcontainers/path"
	"github.com/aidar-khabibullin/MSA_Rocket_Factory/platform/pkg/testcontainers/postgres"
)

const (
	// Параметры для контейнеров
	orderAppName    = "order-app"
	orderDockerfile = "deploy/docker/order/Dockerfile"

	// Переменные окружения приложения
	httpPortKey = "HTTP_PORT"

	// Значения переменных окружения
	loggerLevelValue = "debug"
	startupTimeout   = 3 * time.Minute
)

// setupTestEnvironment — подготавливает тестовое окружение: сеть, контейнеры и возвращает структуру с ресурсами
func setupTestEnvironment(ctx context.Context) *TestEnvironment {
	logger.Info(ctx, "🚀 Подготовка тестового окружения...")

	// Шаг 1: Создаём общую Docker-сеть
	generatedNetwork, err := network.NewNetwork(ctx, "order-service")
	if err != nil {
		logger.Fatal(ctx, "не удалось создать общую сеть", zap.Error(err))
	}
	logger.Info(ctx, "✅ Сеть успешно создана")

	// Получаем переменные окружения для PostgreSQL с проверкой на наличие
	postgresUsername := getEnvWithDefault(ctx, testcontainers.PostgresUsernameKey, testcontainers.PostgresUsername)
	postgresPassword := getEnvWithDefault(ctx, testcontainers.PostgresPasswordKey, testcontainers.PostgresPassword)
	postgresImageName := getEnvWithDefault(ctx, testcontainers.PostgresImageNameKey, testcontainers.PostgresImageName)
	postgresDatabase := getEnvWithDefault(ctx, testcontainers.PostgresDatabaseKey, testcontainers.PostgresDatabase)

	// Получаем порт HTTP для waitStrategy
	httpPort := getEnvWithDefault(ctx, httpPortKey, "8080")

	// Шаг 2: Запускаем контейнер с PostgreSQL
	generatedPostgres, err := postgres.NewContainer(ctx,
		postgres.WithNetworkName(generatedNetwork.Name()),
		postgres.WithContainerName(testcontainers.PostgresContainerName),
		postgres.WithImageName(postgresImageName),
		postgres.WithDatabase(postgresDatabase),
		postgres.WithAuth(postgresUsername, postgresPassword),
		// postgres.WithLogger(logger.Logger()), // Пока отключим логгер
	)
	if err != nil {
		cleanupTestEnvironment(ctx, &TestEnvironment{Network: generatedNetwork})
		logger.Fatal(ctx, "не удалось запустить контейнер PostgreSQL", zap.Error(err))
	}
	logger.Info(ctx, "✅ Контейнер PostgreSQL успешно запущен")

	// Шаг 3: Запускаем контейнер с приложением
	projectRoot := path.GetProjectRoot()

	// Создаем полный набор переменных окружения для приложения
	appEnv := map[string]string{
		// Настройки HTTP
		"HTTP_HOST": "0.0.0.0",
		"HTTP_PORT": httpPort,

		// Настройки логгера
		"LOGGER_LEVEL":   "debug",
		"LOGGER_AS_JSON": "false",

		// Настройки PostgreSQL - используем значения из конфигурации PostgreSQL контейнера
		testcontainers.PostgresHostKey:     generatedPostgres.Config().ContainerName,
		testcontainers.PostgresPortKey:     "5432",
		testcontainers.PostgresDatabaseKey: generatedPostgres.Config().Database,
		testcontainers.PostgresUsernameKey: generatedPostgres.Config().Username,
		testcontainers.PostgresPasswordKey: generatedPostgres.Config().Password,

		// Настройки gRPC клиентов (заглушки для тестов)
		"GRPC_INVENTORY_HOST": "localhost",
		"GRPC_INVENTORY_PORT": "50051",
		"GRPC_PAYMENT_HOST":   "localhost",
		"GRPC_PAYMENT_PORT":   "50052",
	}

	// Создаем настраиваемую стратегию ожидания с увеличенным таймаутом
	waitStrategy := wait.ForListeningPort(nat.Port(httpPort + "/tcp")).
		WithStartupTimeout(startupTimeout)

	appContainer, err := app.NewContainer(ctx,
		app.WithName(orderAppName),
		app.WithPort(httpPort),
		app.WithDockerfile(projectRoot, orderDockerfile),
		app.WithNetwork(generatedNetwork.Name()),
		app.WithEnv(appEnv),
		app.WithLogOutput(os.Stdout),
		app.WithStartupWait(waitStrategy),
		app.WithLogger(logger.Logger()),
	)
	if err != nil {
		cleanupTestEnvironment(ctx, &TestEnvironment{Network: generatedNetwork, Postgres: generatedPostgres})
		logger.Fatal(ctx, "не удалось запустить контейнер приложения", zap.Error(err))
	}
	logger.Info(ctx, "✅ Контейнер приложения успешно запущен")

	logger.Info(ctx, "🎉 Тестовое окружение готово")
	return &TestEnvironment{
		Network:  generatedNetwork,
		Postgres: generatedPostgres,
		App:      appContainer,
	}
}

// getEnvWithLogging возвращает значение переменной окружения с логированием
func getEnvWithLogging(ctx context.Context, key string) string {
	value := os.Getenv(key)
	if value == "" {
		logger.Warn(ctx, "Переменная окружения не установлена", zap.String("key", key))
	}

	return value
}

// getEnvWithDefault возвращает значение переменной окружения или значение по умолчанию
func getEnvWithDefault(ctx context.Context, key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		logger.Info(ctx, "Используется значение по умолчанию для переменной окружения",
			zap.String("key", key), zap.String("default", defaultValue))
		return defaultValue
	}

	return value
}

// cleanupTestEnvironment — вспомогательная функция для освобождения ресурсов
func cleanupTestEnvironment(ctx context.Context, env *TestEnvironment) {
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
