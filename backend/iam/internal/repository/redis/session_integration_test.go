//go:build integration

package redis_test

import (
	"context"
	"testing"
	"time"

	redigo "github.com/gomodule/redigo/redis"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/zap"

	"github.com/aidar-khabibullin/MSA_Rocket_Factory/iam/internal/model"
	sessionRepo "github.com/aidar-khabibullin/MSA_Rocket_Factory/iam/internal/repository/redis"
	cacheRedis "github.com/aidar-khabibullin/MSA_Rocket_Factory/platform/pkg/cache/redis"
)

type nopLogger struct{}

func (nopLogger) Info(context.Context, string, ...zap.Field)  {}
func (nopLogger) Error(context.Context, string, ...zap.Field) {}

// startRedis поднимает настоящий Redis и возвращает репозиторий сессий и сырое соединение для проверки TTL.
func startRedis(t *testing.T) (repo interface {
	GetSessionByUUID(context.Context, uuid.UUID) (*model.Session, error)
	Set(context.Context, uuid.UUID, *model.Session, time.Duration) error
	Delete(context.Context, uuid.UUID) error
}, raw *redigo.Pool) {
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

	pool := &redigo.Pool{
		MaxIdle: 2,
		Dial:    func() (redigo.Conn, error) { return redigo.Dial("tcp", endpoint) },
	}
	t.Cleanup(func() { _ = pool.Close() })

	return sessionRepo.NewRepository(cacheRedis.NewClient(pool, nopLogger{}, 2*time.Second)), pool
}

func newSession() *model.Session {
	now := time.Now().UTC().Truncate(time.Second)
	return &model.Session{UUID: uuid.New(), UserUUID: uuid.New(), CreatedAt: now, UpdatedAt: now, ExpiresAt: now.Add(time.Hour)}
}

func TestRedisSessionRepository(t *testing.T) {
	repo, pool := startRedis(t)
	ctx := context.Background()

	t.Run("сохранённая сессия читается без потерь", func(t *testing.T) {
		session := newSession()
		require.NoError(t, repo.Set(ctx, session.UUID, session, time.Hour))

		got, err := repo.GetSessionByUUID(ctx, session.UUID)

		require.NoError(t, err)
		require.Equal(t, session.UUID, got.UUID)
		require.Equal(t, session.UserUUID, got.UserUUID)
		require.True(t, session.ExpiresAt.Equal(got.ExpiresAt))
	})

	t.Run("ключ получает TTL: сессии не копятся в Redis навсегда", func(t *testing.T) {
		session := newSession()
		require.NoError(t, repo.Set(ctx, session.UUID, session, 90*time.Second))

		conn := pool.Get()
		defer conn.Close()
		ttl, err := redigo.Int(conn.Do("TTL", "iam:session:"+session.UUID.String()))

		require.NoError(t, err)
		require.Greater(t, ttl, 0, "TTL не должен быть -1 (без срока)")
		require.LessOrEqual(t, ttl, 90)
	})

	t.Run("отсутствующая сессия даёт ErrSessionNotFound", func(t *testing.T) {
		_, err := repo.GetSessionByUUID(ctx, uuid.New())

		require.ErrorIs(t, err, model.ErrSessionNotFound)
	})

	t.Run("удалённая сессия больше не читается", func(t *testing.T) {
		session := newSession()
		require.NoError(t, repo.Set(ctx, session.UUID, session, time.Hour))

		require.NoError(t, repo.Delete(ctx, session.UUID))

		_, err := repo.GetSessionByUUID(ctx, session.UUID)
		require.ErrorIs(t, err, model.ErrSessionNotFound)
	})

	t.Run("сессия исчезает по истечении TTL", func(t *testing.T) {
		session := newSession()
		require.NoError(t, repo.Set(ctx, session.UUID, session, time.Second))

		require.Eventually(t, func() bool {
			_, err := repo.GetSessionByUUID(ctx, session.UUID)
			return err != nil
		}, 5*time.Second, 200*time.Millisecond)
	})
}
