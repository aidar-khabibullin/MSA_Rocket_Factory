package cache

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/aidar-khabibullin/MSA_Rocket_Factory/iam/internal/model"
	"github.com/aidar-khabibullin/MSA_Rocket_Factory/iam/internal/repository"
	"github.com/aidar-khabibullin/MSA_Rocket_Factory/platform/pkg/logger"
)

// CacheManager управляет кешированием с дополнительной логикой
type CacheManager struct {
	sessionCache repository.SessionCache
	mu           sync.RWMutex

	// Статистика кеша
	stats CacheStats
}

// CacheStats содержит статистику работы кеша
type CacheStats struct {
	Hits     int64
	Misses   int64
	Errors   int64
	Sets     int64
	Deletes  int64
	LastHit  time.Time
	LastMiss time.Time
}

// NewCacheManager создает новый менеджер кеша
func NewCacheManager(sessionCache repository.SessionCache) *CacheManager {
	return &CacheManager{
		sessionCache: sessionCache,
		stats:        CacheStats{},
	}
}

// GetSession получает сессию из кеша с логированием статистики
func (cm *CacheManager) GetSession(ctx context.Context, sessionUUID uuid.UUID) (*model.Session, error) {
	session, err := cm.sessionCache.GetSessionByUUID(ctx, sessionUUID)

	cm.mu.Lock()
	defer cm.mu.Unlock()

	if err != nil {
		if errors.Is(err, model.ErrSessionNotFound) {
			cm.stats.Misses++
			cm.stats.LastMiss = time.Now()
			logger.Debug(ctx, "🔍 Cache miss для сессии",
				zap.String("session_uuid", sessionUUID.String()))
		} else {
			cm.stats.Errors++
			logger.Error(ctx, "❌ Ошибка получения сессии из кеша",
				zap.String("session_uuid", sessionUUID.String()),
				zap.Error(err))
		}
		return nil, err
	}

	cm.stats.Hits++
	cm.stats.LastHit = time.Now()
	logger.Debug(ctx, "✅ Cache hit для сессии",
		zap.String("session_uuid", sessionUUID.String()))

	return session, nil
}

// SetSession сохраняет сессию в кеш с TTL
func (cm *CacheManager) SetSession(ctx context.Context, sessionUUID uuid.UUID, session *model.Session, ttl time.Duration) error {
	err := cm.sessionCache.Set(ctx, sessionUUID, session, ttl)

	cm.mu.Lock()
	defer cm.mu.Unlock()

	if err != nil {
		cm.stats.Errors++
		logger.Error(ctx, "❌ Ошибка сохранения сессии в кеш",
			zap.String("session_uuid", sessionUUID.String()),
			zap.Duration("ttl", ttl),
			zap.Error(err))
		return err
	}

	cm.stats.Sets++
	logger.Debug(ctx, "💾 Сессия сохранена в кеш",
		zap.String("session_uuid", sessionUUID.String()),
		zap.Duration("ttl", ttl))

	return nil
}

// DeleteSession удаляет сессию из кеша
func (cm *CacheManager) DeleteSession(ctx context.Context, sessionUUID uuid.UUID) error {
	err := cm.sessionCache.Delete(ctx, sessionUUID)

	cm.mu.Lock()
	defer cm.mu.Unlock()

	if err != nil {
		cm.stats.Errors++
		logger.Error(ctx, "❌ Ошибка удаления сессии из кеша",
			zap.String("session_uuid", sessionUUID.String()),
			zap.Error(err))
		return err
	}

	cm.stats.Deletes++
	logger.Debug(ctx, "🗑️ Сессия удалена из кеша",
		zap.String("session_uuid", sessionUUID.String()))

	return nil
}

// GetStats возвращает статистику кеша
func (cm *CacheManager) GetStats() CacheStats {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	return cm.stats
}

// GetHitRatio вычисляет процент попаданий в кеш
func (cm *CacheManager) GetHitRatio() float64 {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	total := cm.stats.Hits + cm.stats.Misses
	if total == 0 {
		return 0.0
	}

	return float64(cm.stats.Hits) / float64(total) * 100
}

// LogStats логирует статистику кеша
func (cm *CacheManager) LogStats(ctx context.Context) {
	stats := cm.GetStats()
	hitRatio := cm.GetHitRatio()

	logger.Info(ctx, "📈 Статистика кеша",
		zap.Int64("hits", stats.Hits),
		zap.Int64("misses", stats.Misses),
		zap.Int64("errors", stats.Errors),
		zap.Int64("sets", stats.Sets),
		zap.Int64("deletes", stats.Deletes),
		zap.Float64("hit_ratio_percent", hitRatio),
		zap.Time("last_hit", stats.LastHit),
		zap.Time("last_miss", stats.LastMiss),
	)
}

// ResetStats сбрасывает статистику кеша
func (cm *CacheManager) ResetStats() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.stats = CacheStats{}
}

// WarmupCache предварительно прогревает кеш часто используемыми данными
func (cm *CacheManager) WarmupCache(ctx context.Context, sessions []*model.Session) error {
	logger.Info(ctx, "🔥 Начинаем прогрев кеша", zap.Int("sessions_count", len(sessions)))

	var errors []error
	warmedCount := 0

	for _, session := range sessions {
		if session.IsExpired() {
			continue // Пропускаем истекшие сессии
		}

		ttl := time.Until(session.ExpiresAt)
		if ttl > 0 {
			if err := cm.SetSession(ctx, session.UUID, session, ttl); err != nil {
				errors = append(errors, fmt.Errorf("не удалось прогреть сессию %s: %w", session.UUID.String(), err))
			} else {
				warmedCount++
			}
		}
	}

	logger.Info(ctx, "✅ Прогрев кеша завершен",
		zap.Int("warmed_sessions", warmedCount),
		zap.Int("total_sessions", len(sessions)),
		zap.Int("errors", len(errors)))

	if len(errors) > 0 {
		for _, err := range errors {
			logger.Warn(ctx, "⚠️ Ошибка прогрева кеша", zap.Error(err))
		}
		return fmt.Errorf("прогрев кеша завершен с %d ошибками", len(errors))
	}

	return nil
}
