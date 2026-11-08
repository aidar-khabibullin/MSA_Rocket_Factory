package composite_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/aidar-khabibullin/MSA_Rocket_Factory/iam/internal/model"
	"github.com/aidar-khabibullin/MSA_Rocket_Factory/iam/internal/repository/composite"
	"github.com/aidar-khabibullin/MSA_Rocket_Factory/iam/internal/repository/mocks"
)

// migratingRepo добавляет к мок-репозиторию метод Migrate, как у настоящего PostgreSQL-репозитория
type migratingRepo struct {
	*mocks.IAMRepository
	migrateErr error
	gotDir     string
}

func (m *migratingRepo) Migrate(dir string) error {
	m.gotDir = dir
	return m.migrateErr
}

func TestUserMethodsGoToPostgres(t *testing.T) {
	ctx := context.Background()
	pg, cache := mocks.NewIAMRepository(t), mocks.NewIAMRepository(t)
	repo := composite.NewRepository(pg, cache)
	user := &model.User{UUID: uuid.New(), Login: "rocket"}

	pg.On("CreateUser", mock.Anything, user).Return(user, nil)
	pg.On("GetUserByUUID", mock.Anything, user.UUID).Return(user, nil)
	pg.On("GetUserByLogin", mock.Anything, "rocket").Return(user, nil)
	pg.On("UpdateUser", mock.Anything, user).Return(user, nil)

	got, err := repo.CreateUser(ctx, user)
	require.NoError(t, err)
	require.Equal(t, user, got)

	got, err = repo.GetUserByUUID(ctx, user.UUID)
	require.NoError(t, err)
	require.Equal(t, user, got)

	got, err = repo.GetUserByLogin(ctx, "rocket")
	require.NoError(t, err)
	require.Equal(t, user, got)

	got, err = repo.UpdateUser(ctx, user)
	require.NoError(t, err)
	require.Equal(t, user, got)
}

func TestSessionMethodsGoToPostgres(t *testing.T) {
	ctx := context.Background()
	pg, cache := mocks.NewIAMRepository(t), mocks.NewIAMRepository(t)
	repo := composite.NewRepository(pg, cache)
	session := &model.Session{UUID: uuid.New()}
	dbErr := errors.New("db error")

	pg.On("CreateSession", mock.Anything, session).Return(session, nil)
	pg.On("GetSessionByUUID", mock.Anything, session.UUID).Return(session, nil)
	pg.On("UpdateSession", mock.Anything, session).Return(nil, dbErr)
	pg.On("DeleteSession", mock.Anything, session.UUID).Return(nil)
	pg.On("CleanupExpiredSessions", mock.Anything).Return(nil)

	got, err := repo.CreateSession(ctx, session)
	require.NoError(t, err)
	require.Equal(t, session, got)

	got, err = repo.GetSessionByUUID(ctx, session.UUID)
	require.NoError(t, err)
	require.Equal(t, session, got)

	_, err = repo.UpdateSession(ctx, session)
	require.ErrorIs(t, err, dbErr)

	require.NoError(t, repo.DeleteSession(ctx, session.UUID))
	require.NoError(t, repo.CleanupExpiredSessions(ctx))
}

func TestCacheMethodsGoToCache(t *testing.T) {
	ctx := context.Background()
	pg, cache := mocks.NewIAMRepository(t), mocks.NewIAMRepository(t)
	repo := composite.NewRepository(pg, cache)
	session := &model.Session{UUID: uuid.New()}

	cache.On("GetSessionByUUID", mock.Anything, session.UUID).Return(session, nil)
	cache.On("Set", mock.Anything, session.UUID, session, time.Minute).Return(nil)
	cache.On("Delete", mock.Anything, session.UUID).Return(nil)

	got, err := repo.GetSessionFromCache(ctx, session.UUID)
	require.NoError(t, err)
	require.Equal(t, session, got)
	require.NoError(t, repo.Set(ctx, session.UUID, session, time.Minute))
	require.NoError(t, repo.Delete(ctx, session.UUID))
}

func TestMigrate(t *testing.T) {
	t.Run("unsupported by postgres repository", func(t *testing.T) {
		repo := composite.NewRepository(mocks.NewIAMRepository(t), mocks.NewIAMRepository(t))

		require.Error(t, repo.(interface{ Migrate(string) error }).Migrate("migrations"))
	})

	t.Run("delegated to postgres repository", func(t *testing.T) {
		pg := &migratingRepo{IAMRepository: mocks.NewIAMRepository(t)}
		repo := composite.NewRepository(pg, mocks.NewIAMRepository(t))

		require.NoError(t, repo.(interface{ Migrate(string) error }).Migrate("migrations"))
		require.Equal(t, "migrations", pg.gotDir)
	})

	t.Run("migration error is propagated", func(t *testing.T) {
		migrateErr := errors.New("migrate failed")
		pg := &migratingRepo{IAMRepository: mocks.NewIAMRepository(t), migrateErr: migrateErr}
		repo := composite.NewRepository(pg, mocks.NewIAMRepository(t))

		require.ErrorIs(t, repo.(interface{ Migrate(string) error }).Migrate("migrations"), migrateErr)
	})
}
