package cache_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/aidar-khabibullin/MSA_Rocket_Factory/iam/internal/model"
	"github.com/aidar-khabibullin/MSA_Rocket_Factory/iam/internal/repository/mocks"
	"github.com/aidar-khabibullin/MSA_Rocket_Factory/iam/internal/service/cache"
	"github.com/aidar-khabibullin/MSA_Rocket_Factory/platform/pkg/logger"
)

func newManager(t *testing.T) (*cache.CacheManager, *mocks.IAMRepository) {
	t.Helper()
	logger.SetNopLogger()

	repo := mocks.NewIAMRepository(t)

	return cache.NewCacheManager(repo), repo
}

func TestGetSession_HitMissAndError(t *testing.T) {
	ctx := context.Background()
	cm, repo := newManager(t)
	hit, miss, boom := uuid.New(), uuid.New(), uuid.New()
	session := &model.Session{UUID: hit}

	repo.On("GetSessionByUUID", mock.Anything, hit).Return(session, nil)
	repo.On("GetSessionByUUID", mock.Anything, miss).Return(nil, model.ErrSessionNotFound)
	repo.On("GetSessionByUUID", mock.Anything, boom).Return(nil, errors.New("redis down"))

	got, err := cm.GetSession(ctx, hit)
	require.NoError(t, err)
	require.Equal(t, session, got)

	_, err = cm.GetSession(ctx, miss)
	require.ErrorIs(t, err, model.ErrSessionNotFound)

	_, err = cm.GetSession(ctx, boom)
	require.Error(t, err)

	stats := cm.GetStats()
	require.EqualValues(t, 1, stats.Hits)
	require.EqualValues(t, 1, stats.Misses)
	require.EqualValues(t, 1, stats.Errors)
	require.False(t, stats.LastHit.IsZero())
	require.False(t, stats.LastMiss.IsZero())
	require.InDelta(t, 50.0, cm.GetHitRatio(), 0.001)
}

func TestHitRatio_NoRequests(t *testing.T) {
	cm, _ := newManager(t)

	require.Zero(t, cm.GetHitRatio())
}

func TestSetAndDeleteSession(t *testing.T) {
	ctx := context.Background()
	cm, repo := newManager(t)
	id := uuid.New()
	session := &model.Session{UUID: id}
	failing := uuid.New()

	repo.On("Set", mock.Anything, id, session, time.Minute).Return(nil)
	repo.On("Set", mock.Anything, failing, session, time.Minute).Return(errors.New("set failed"))
	repo.On("Delete", mock.Anything, id).Return(nil)
	repo.On("Delete", mock.Anything, failing).Return(errors.New("delete failed"))

	require.NoError(t, cm.SetSession(ctx, id, session, time.Minute))
	require.Error(t, cm.SetSession(ctx, failing, session, time.Minute))
	require.NoError(t, cm.DeleteSession(ctx, id))
	require.Error(t, cm.DeleteSession(ctx, failing))

	stats := cm.GetStats()
	require.EqualValues(t, 1, stats.Sets)
	require.EqualValues(t, 1, stats.Deletes)
	require.EqualValues(t, 2, stats.Errors)
}

func TestLogAndResetStats(t *testing.T) {
	ctx := context.Background()
	cm, repo := newManager(t)
	id := uuid.New()
	repo.On("Delete", mock.Anything, id).Return(nil)
	require.NoError(t, cm.DeleteSession(ctx, id))

	cm.LogStats(ctx)
	cm.ResetStats()

	require.Equal(t, cache.CacheStats{}, cm.GetStats())
}

func TestWarmupCache(t *testing.T) {
	ctx := context.Background()
	future := time.Now().Add(time.Hour)

	t.Run("skips expired sessions and caches the rest", func(t *testing.T) {
		cm, repo := newManager(t)
		valid := &model.Session{UUID: uuid.New(), ExpiresAt: future}
		expired := &model.Session{UUID: uuid.New(), ExpiresAt: time.Now().Add(-time.Hour)}
		repo.On("Set", mock.Anything, valid.UUID, valid, mock.AnythingOfType("time.Duration")).Return(nil).Once()

		require.NoError(t, cm.WarmupCache(ctx, []*model.Session{valid, expired}))
		require.EqualValues(t, 1, cm.GetStats().Sets)
	})

	t.Run("reports cache errors", func(t *testing.T) {
		cm, repo := newManager(t)
		valid := &model.Session{UUID: uuid.New(), ExpiresAt: future}
		repo.On("Set", mock.Anything, valid.UUID, valid, mock.AnythingOfType("time.Duration")).
			Return(errors.New("set failed")).Once()

		require.Error(t, cm.WarmupCache(ctx, []*model.Session{valid}))
	})

	t.Run("empty input", func(t *testing.T) {
		cm, _ := newManager(t)

		require.NoError(t, cm.WarmupCache(ctx, nil))
	})
}
