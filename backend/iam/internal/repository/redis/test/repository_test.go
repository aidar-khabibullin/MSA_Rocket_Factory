package redis_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	redigo "github.com/gomodule/redigo/redis"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/aidar-khabibullin/MSA_Rocket_Factory/iam/internal/model"
	"github.com/aidar-khabibullin/MSA_Rocket_Factory/iam/internal/repository/redis"
	"github.com/aidar-khabibullin/MSA_Rocket_Factory/platform/pkg/cache"
)

// fakeRedis - in-memory реализация нужных репозиторию методов RedisClient.
// Hash хранится так же, как возвращает redigo: плоский список "поле, значение" в []byte.
type fakeRedis struct {
	cache.RedisClient

	hashes  map[string][]any
	ttls    map[string]time.Duration
	hGetErr error
	hSetErr error
	expErr  error
	delErr  error
}

func newFakeRedis() *fakeRedis {
	return &fakeRedis{hashes: map[string][]any{}, ttls: map[string]time.Duration{}}
}

func (f *fakeRedis) HashSet(_ context.Context, key string, values any) error {
	if f.hSetErr != nil {
		return f.hSetErr
	}

	args := redigo.Args{}.AddFlat(values)
	flat := make([]any, 0, len(args))
	for _, a := range args {
		flat = append(flat, []byte(fmt.Sprint(a)))
	}
	f.hashes[key] = flat

	return nil
}

func (f *fakeRedis) HGetAll(_ context.Context, key string) ([]any, error) {
	if f.hGetErr != nil {
		return nil, f.hGetErr
	}

	return f.hashes[key], nil
}

func (f *fakeRedis) Expire(_ context.Context, key string, ttl time.Duration) error {
	if f.expErr != nil {
		return f.expErr
	}
	f.ttls[key] = ttl

	return nil
}

func (f *fakeRedis) Del(_ context.Context, key string) error {
	if f.delErr != nil {
		return f.delErr
	}
	delete(f.hashes, key)

	return nil
}

func testSession() *model.Session {
	now := time.Now().Truncate(time.Second)

	return &model.Session{
		UUID:      uuid.New(),
		UserUUID:  uuid.New(),
		CreatedAt: now,
		UpdatedAt: now.Add(time.Minute),
		ExpiresAt: now.Add(time.Hour),
	}
}

func TestSetThenGetRoundTrip(t *testing.T) {
	ctx := context.Background()
	fake := newFakeRedis()
	repo := redis.NewRepository(fake)
	session := testSession()

	require.NoError(t, repo.Set(ctx, session.UUID, session, 30*time.Minute))

	got, err := repo.GetSessionByUUID(ctx, session.UUID)
	require.NoError(t, err)
	require.Equal(t, session.UUID, got.UUID)
	require.Equal(t, session.UserUUID, got.UserUUID)
	require.True(t, session.ExpiresAt.Equal(got.ExpiresAt))
	require.Equal(t, 30*time.Minute, fake.ttls["iam:session:"+session.UUID.String()])
}

func TestGetSession_NotFound(t *testing.T) {
	repo := redis.NewRepository(newFakeRedis())

	_, err := repo.GetSessionByUUID(context.Background(), uuid.New())

	require.ErrorIs(t, err, model.ErrSessionNotFound)
}

func TestGetSession_NilErrorMapsToNotFound(t *testing.T) {
	fake := newFakeRedis()
	fake.hGetErr = redigo.ErrNil
	repo := redis.NewRepository(fake)

	_, err := repo.GetSessionByUUID(context.Background(), uuid.New())

	require.ErrorIs(t, err, model.ErrSessionNotFound)
}

func TestGetSession_ClientError(t *testing.T) {
	fake := newFakeRedis()
	fake.hGetErr = errors.New("connection refused")
	repo := redis.NewRepository(fake)

	_, err := repo.GetSessionByUUID(context.Background(), uuid.New())

	require.EqualError(t, err, "connection refused")
}

func TestGetSession_CorruptedData(t *testing.T) {
	ctx := context.Background()
	id := uuid.New()

	t.Run("odd number of fields", func(t *testing.T) {
		fake := newFakeRedis()
		fake.hashes["iam:session:"+id.String()] = []any{[]byte("uuid")}

		_, err := redis.NewRepository(fake).GetSessionByUUID(ctx, id)

		require.Error(t, err)
	})

	t.Run("invalid uuid in hash", func(t *testing.T) {
		fake := newFakeRedis()
		fake.hashes["iam:session:"+id.String()] = []any{
			[]byte("uuid"), []byte("not-a-uuid"),
			[]byte("user_uuid"), []byte(uuid.NewString()),
		}

		_, err := redis.NewRepository(fake).GetSessionByUUID(ctx, id)

		require.Error(t, err)
	})
}

func TestSet_Errors(t *testing.T) {
	ctx := context.Background()
	session := testSession()

	t.Run("hash set fails", func(t *testing.T) {
		fake := newFakeRedis()
		fake.hSetErr = errors.New("hset failed")

		require.EqualError(t, redis.NewRepository(fake).Set(ctx, session.UUID, session, time.Minute), "hset failed")
	})

	t.Run("expire fails", func(t *testing.T) {
		fake := newFakeRedis()
		fake.expErr = errors.New("expire failed")

		require.EqualError(t, redis.NewRepository(fake).Set(ctx, session.UUID, session, time.Minute), "expire failed")
	})
}

func TestDelete(t *testing.T) {
	ctx := context.Background()
	fake := newFakeRedis()
	repo := redis.NewRepository(fake)
	session := testSession()
	require.NoError(t, repo.Set(ctx, session.UUID, session, time.Minute))

	require.NoError(t, repo.Delete(ctx, session.UUID))
	_, err := repo.GetSessionByUUID(ctx, session.UUID)
	require.ErrorIs(t, err, model.ErrSessionNotFound)

	fake.delErr = errors.New("del failed")
	require.EqualError(t, repo.Delete(ctx, session.UUID), "del failed")
}
