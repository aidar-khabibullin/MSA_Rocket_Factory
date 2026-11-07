//go:build integration

package mongo_test

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	mongodriver "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/aidar-khabibullin/MSA_Rocket_Factory/inventory/internal/model"
	"github.com/aidar-khabibullin/MSA_Rocket_Factory/inventory/internal/repository"
	mongoRepo "github.com/aidar-khabibullin/MSA_Rocket_Factory/inventory/internal/repository/mongo"
)

// Тестовые данные из mongo/testdata.go: RD-180 (остаток 5) и Крыло Delta-X (остаток 4).
var (
	engine = uuid.MustParse("d973e963-b7e6-4323-8f4e-4bfd5ab8e834")
	wing   = uuid.MustParse("d973e963-b7e6-4323-8f4e-4bfd5ab8e838")
)

func newRepo(t *testing.T) repository.InventoryRepository {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	t.Cleanup(cancel)

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "mongo:7.0.5",
			ExposedPorts: []string{"27017/tcp"},
			WaitingFor:   wait.ForLog("Waiting for connections").WithStartupTimeout(90 * time.Second),
		},
		Started: true,
	})
	require.NoError(t, err)
	t.Cleanup(func() { _ = container.Terminate(context.Background()) })

	endpoint, err := container.Endpoint(ctx, "")
	require.NoError(t, err)

	client, err := mongodriver.Connect(ctx, options.Client().ApplyURI("mongodb://"+endpoint))
	require.NoError(t, err)
	t.Cleanup(func() { _ = client.Disconnect(context.Background()) })

	return mongoRepo.NewRepository(ctx, client.Database("inventory-test"))
}

func stock(t *testing.T, repo repository.InventoryRepository, id uuid.UUID) int64 {
	t.Helper()
	part, err := repo.GetPart(context.Background(), id)
	require.NoError(t, err)
	return part.StockQuantity
}

func TestMongoReserve(t *testing.T) {
	repo := newRepo(t)
	ctx := context.Background()

	t.Run("списывает остаток, возврат восстанавливает его ровно один раз", func(t *testing.T) {
		order := uuid.New()
		require.NoError(t, repo.Reserve(ctx, order, []uuid.UUID{engine, engine}))
		require.Equal(t, int64(3), stock(t, repo, engine))

		require.NoError(t, repo.Release(ctx, order))
		require.NoError(t, repo.Release(ctx, order))
		require.Equal(t, int64(5), stock(t, repo, engine))
	})

	t.Run("идемпотентен по заказу", func(t *testing.T) {
		order := uuid.New()
		require.NoError(t, repo.Reserve(ctx, order, []uuid.UUID{engine}))
		require.NoError(t, repo.Reserve(ctx, order, []uuid.UUID{engine}))
		require.Equal(t, int64(4), stock(t, repo, engine))
		require.NoError(t, repo.Release(ctx, order))
	})

	t.Run("при нехватке одной детали откатывает уже списанные", func(t *testing.T) {
		// engine доступен, но крыльев нужно 5 при остатке 4.
		parts := []uuid.UUID{engine, wing, wing, wing, wing, wing}

		err := repo.Reserve(ctx, uuid.New(), parts)

		require.ErrorIs(t, err, model.ErrInsufficientStock)
		require.Equal(t, int64(5), stock(t, repo, engine), "списание engine должно откатиться")
		require.Equal(t, int64(4), stock(t, repo, wing))
	})

	t.Run("неизвестная деталь", func(t *testing.T) {
		require.ErrorIs(t, repo.Reserve(ctx, uuid.New(), []uuid.UUID{engine, uuid.New()}), model.ErrPartNotFound)
		require.Equal(t, int64(5), stock(t, repo, engine))
	})

	t.Run("параллельные заказы не продают больше остатка", func(t *testing.T) {
		var ok atomic.Int64
		var wg sync.WaitGroup
		for range 20 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				if err := repo.Reserve(ctx, uuid.New(), []uuid.UUID{wing}); err == nil {
					ok.Add(1)
				} else {
					require.ErrorIs(t, err, model.ErrInsufficientStock)
				}
			}()
		}
		wg.Wait()

		require.Equal(t, int64(4), ok.Load())
		require.Equal(t, int64(0), stock(t, repo, wing))
	})

	t.Run("параллельный резерв одного заказа списывает остаток не более одного раза", func(t *testing.T) {
		order := uuid.New()
		before := stock(t, repo, engine)

		var wg sync.WaitGroup
		for range 10 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				// Параллельные дубликаты при тесном остатке могут получить ErrInsufficientStock: каждый успевает
				// списать остаток до проверки уникального индекса. Важно, что итоговое списание одно.
				if err := repo.Reserve(ctx, order, []uuid.UUID{engine}); err != nil {
					require.ErrorIs(t, err, model.ErrInsufficientStock)
				}
			}()
		}
		wg.Wait()

		require.Equal(t, before-1, stock(t, repo, engine), "остаток списан ровно один раз")
		require.NoError(t, repo.Reserve(ctx, order, []uuid.UUID{engine}), "последовательный повтор идемпотентен")
		require.Equal(t, before-1, stock(t, repo, engine))
	})
}
