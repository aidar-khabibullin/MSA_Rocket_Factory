package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel"

	"github.com/aidar-khabibullin/MSA_Rocket_Factory/order/internal/api/health"
	customMiddleware "github.com/aidar-khabibullin/MSA_Rocket_Factory/order/internal/api/middleware"
	"github.com/aidar-khabibullin/MSA_Rocket_Factory/order/internal/config"
	"github.com/aidar-khabibullin/MSA_Rocket_Factory/platform/pkg/closer"
	"github.com/aidar-khabibullin/MSA_Rocket_Factory/platform/pkg/logger"
	"github.com/aidar-khabibullin/MSA_Rocket_Factory/platform/pkg/metrics"
	platformHTTPMiddleware "github.com/aidar-khabibullin/MSA_Rocket_Factory/platform/pkg/middleware/http"
	"github.com/aidar-khabibullin/MSA_Rocket_Factory/platform/pkg/runner"
	"github.com/aidar-khabibullin/MSA_Rocket_Factory/platform/pkg/tracing"
	orderV1 "github.com/aidar-khabibullin/MSA_Rocket_Factory/shared/pkg/openapi/order/v1"
)

type App struct {
	diContainer *diContainer
	httpServer  *http.Server
}

// extractDBName извлекает имя базы данных из URI
func extractDBName(uri string) (string, error) {
	if len(uri) == 0 {
		return "", fmt.Errorf("пустой URI")
	}

	lastSlash := -1
	questionMark := -1
	for i := len(uri) - 1; i >= 0; i-- {
		if uri[i] == '?' && questionMark == -1 {
			questionMark = i
		}
		if uri[i] == '/' && lastSlash == -1 {
			lastSlash = i
			break
		}
	}

	if lastSlash == -1 || questionMark == -1 || lastSlash >= questionMark {
		return "", fmt.Errorf("неверный формат URI")
	}

	return uri[lastSlash+1 : questionMark], nil
}

// createSystemURI создает URI для подключения к системной БД postgres
func createSystemURI(uri string) (string, error) {
	if len(uri) == 0 {
		return "", fmt.Errorf("пустой URI")
	}

	lastSlash := -1
	questionMark := -1
	for i := len(uri) - 1; i >= 0; i-- {
		if uri[i] == '?' && questionMark == -1 {
			questionMark = i
		}
		if uri[i] == '/' && lastSlash == -1 {
			lastSlash = i
			break
		}
	}

	if lastSlash == -1 || questionMark == -1 || lastSlash >= questionMark {
		return "", fmt.Errorf("неверный формат URI")
	}

	return uri[:lastSlash+1] + "postgres" + uri[questionMark:], nil
}

// checkAndCreateDB проверяет существование БД и создает её при необходимости
func checkAndCreateDB(ctx context.Context, systemURI, dbName string) error {
	pool, err := pgxpool.New(ctx, systemURI)
	if err != nil {
		return fmt.Errorf("не удалось подключиться к PostgreSQL: %w", err)
	}
	defer pool.Close()

	var exists bool
	err = pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)", dbName).Scan(&exists)
	if err != nil {
		return fmt.Errorf("ошибка при проверке существования БД: %w", err)
	}

	if !exists {
		logger.Info(ctx, fmt.Sprintf("📝 База данных %s не существует, создаем...", dbName))
		_, err = pool.Exec(ctx, fmt.Sprintf("CREATE DATABASE \"%s\"", dbName))
		if err != nil {
			return fmt.Errorf("не удалось создать БД %s: %w", dbName, err)
		}
		logger.Info(ctx, fmt.Sprintf("✅ База данных %s успешно создана", dbName))
	} else {
		logger.Info(ctx, fmt.Sprintf("✅ База данных %s уже существует", dbName))
	}

	return nil
}

// ensureDatabaseExists проверяет существование базы данных и создает её при необходимости
func (a *App) ensureDatabaseExists(ctx context.Context) error {
	logger.Info(ctx, "🔍 Проверяем существование базы данных...")

	dbConfig := config.AppConfig().DB
	uri := dbConfig.URI()

	dbName, err := extractDBName(uri)
	if err != nil {
		return fmt.Errorf("не удалось извлечь имя БД из URI: %s", uri)
	}

	systemURI, err := createSystemURI(uri)
	if err != nil {
		return fmt.Errorf("не удалось создать URI для системной БД")
	}

	return checkAndCreateDB(ctx, systemURI, dbName)
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
	tasks := []runner.Task{
		{Name: "HTTP server", Run: a.runHTTPServer},
	}

	// Kafka Consumer запускается только если не отключен для тестов
	if os.Getenv("SKIP_KAFKA_CONSUMER") != "true" {
		// Без консьюмера заказы никогда не станут ASSEMBLED, поэтому его падение роняет весь сервис,
		// и Docker (restart: unless-stopped) перезапускает его. Раньше ошибка только логировалась,
		// а HTTP-сервер продолжал работать.
		tasks = append(tasks, runner.ConsumerTask("ShipAssembled Consumer", a.diContainer.ShipAssembledConsumer(ctx)))
	}

	return runner.Run(ctx, tasks...)
}

func (a *App) initDeps(ctx context.Context) error {
	inits := []func(context.Context) error{
		a.initLogger,
		a.initTracing,
		a.initMetrics,
	}

	// Проверяем и создаем БД перед инициализацией DI только если не отключено для тестов
	if os.Getenv("SKIP_DB_CHECK") != "true" {
		inits = append(inits, a.ensureDatabaseExists)
	}

	inits = append(inits, []func(context.Context) error{
		a.initDI,
		a.initCloser,
		a.initHTTPServer,
	}...)

	for _, f := range inits {
		err := f(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *App) initDI(ctx context.Context) error {
	logger.Info(ctx, "🔧 Начинаем инициализацию DI контейнера")
	a.diContainer = NewDiContainer()
	logger.Info(ctx, "✅ DI контейнер создан")

	// Инициализируем OrderPaidProducer и ShipAssembledConsumer и устанавливаем их зависимости
	// Это нужно делать после создания DI контейнера, но до запуска сервера
	if os.Getenv("SKIP_KAFKA_CONSUMER") != "true" {
		logger.Info(ctx, "🔧 Создаем OrderPaidProducer")
		_ = a.diContainer.OrderPaidProducer(ctx) // Создаем producer
		logger.Info(ctx, "✅ OrderPaidProducer создан")

		logger.Info(ctx, "🔧 Создаем ShipAssembledConsumer")
		_ = a.diContainer.ShipAssembledConsumer(ctx) // Создаем consumer
		logger.Info(ctx, "✅ ShipAssembledConsumer создан")
	} else {
		logger.Info(ctx, "⏭️ Пропускаем создание Kafka компонентов (SKIP_KAFKA_CONSUMER=true)")
	}

	// Создаем OrderService и устанавливаем зависимости после создания всех компонентов
	logger.Info(ctx, "🔧 Создаем OrderService")
	_ = a.diContainer.OrderService(ctx) // Создаем OrderService
	logger.Info(ctx, "✅ OrderService создан")

	logger.Info(ctx, "🔧 Создаем OrderV1API")
	_ = a.diContainer.OrderV1API(ctx) // Создаем OrderV1API
	logger.Info(ctx, "✅ OrderV1API создан")

	logger.Info(ctx, "🔧 Устанавливаем зависимости")
	a.diContainer.SetupOrderPaidProducer(ctx) // Устанавливаем producer в OrderService
	logger.Info(ctx, "✅ OrderPaidProducer установлен в OrderService")

	a.diContainer.SetupShipAssembledConsumer(ctx) // Устанавливаем OrderService в consumer
	logger.Info(ctx, "✅ OrderService установлен в ShipAssembledConsumer")

	a.diContainer.SetupOrderV1API(ctx) // Устанавливаем OrderService в API
	logger.Info(ctx, "✅ OrderService установлен в OrderV1API")

	logger.Info(ctx, "🎉 Инициализация DI контейнера завершена")
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

func (a *App) initMetrics(ctx context.Context) error {
	return metrics.InitProvider(ctx, config.AppConfig().Metrics)
}

func (a *App) initCloser(_ context.Context) error {
	closer.SetLogger(logger.Logger())

	// Добавляем graceful shutdown для метрик
	closer.AddNamed("Metrics provider", metrics.Shutdown)

	// Добавляем graceful shutdown для tracing
	closer.AddNamed("Tracing provider", tracing.ShutdownTracer)

	return nil
}

func (a *App) initHTTPServer(ctx context.Context) error {
	// Создаем OpenAPI сервер с OpenTelemetry конфигурацией
	orderServer, err := orderV1.NewServer(
		a.diContainer.OrderV1API(ctx),
		orderV1.WithMeterProvider(metrics.GetMeterProvider()),
		orderV1.WithTracerProvider(otel.GetTracerProvider()),
	)
	if err != nil {
		return fmt.Errorf("failed to create OpenAPI server: %w", err)
	}

	// Создаем HTTP метрики
	httpMetrics, err := customMiddleware.NewHTTPMetrics()
	if err != nil {
		return fmt.Errorf("failed to create HTTP metrics: %w", err)
	}

	// Настраиваем роутер
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(10 * time.Second))
	r.Use(platformHTTPMiddleware.TracingMiddleware("order")) // Добавляем трейсинг middleware
	r.Use(customMiddleware.RequestLogger)
	r.Use(customMiddleware.MetricsMiddleware(httpMetrics))

	// Health check без аутентификации: его вызывают пробы живости и шлюз, у них нет сессии пользователя.
	healthHandler := health.NewHealthHandler()
	r.Get("/health", healthHandler.Handle)

	// Аутентификация действует только на API заказов. В chi нельзя вызывать Use после регистрации
	// маршрутов на том же роутере, поэтому защищённая часть вынесена в группу.
	authMiddleware := a.diContainer.AuthMiddleware(ctx)
	r.Group(func(protected chi.Router) {
		protected.Use(authMiddleware.Handle)
		protected.Mount("/", orderServer)
	})

	// Создаем HTTP сервер
	a.httpServer = &http.Server{
		Addr:              config.AppConfig().HTTP.Address(),
		Handler:           r,
		ReadHeaderTimeout: time.Duration(config.AppConfig().HTTP.ReadHeaderTimeout()) * time.Second,
	}

	closer.AddNamed("HTTP server", func(ctx context.Context) error {
		shutdownCtx, cancel := context.WithTimeout(ctx, time.Duration(config.AppConfig().HTTP.ShutdownTimeout())*time.Second)
		defer cancel()

		err := a.httpServer.Shutdown(shutdownCtx)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	})

	return nil
}

func (a *App) runHTTPServer(ctx context.Context) error {
	logger.Info(ctx, fmt.Sprintf("🚀 HTTP Order service server listening on %s", config.AppConfig().HTTP.Address()))

	errCh := make(chan error, 1)
	go func() {
		errCh <- a.httpServer.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}

		return nil
	case <-ctx.Done():
		// Контекст отменён (сигнал или падение соседней задачи) - останавливаем сервер корректно.
		// WithoutCancel нужен, потому что ctx уже отменён и Shutdown сразу бы прервался.
		shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()

		if err := a.httpServer.Shutdown(shutdownCtx); err != nil {
			return err
		}

		<-errCh

		return nil
	}
}
