package auth

import (
	"context"
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/aidar-khabibullin/MSA_Rocket_Factory/iam/internal/model"
	"github.com/aidar-khabibullin/MSA_Rocket_Factory/iam/internal/repository"
	"github.com/aidar-khabibullin/MSA_Rocket_Factory/platform/pkg/logger"
)

// TokenBlacklistService управляет списком отозванных токенов
type TokenBlacklistService struct {
	cacheRepo repository.SessionCache
}

// NewTokenBlacklistService создает новый сервис blacklist токенов
func NewTokenBlacklistService(cacheRepo repository.SessionCache) *TokenBlacklistService {
	return &TokenBlacklistService{
		cacheRepo: cacheRepo,
	}
}

// RevokeToken добавляет токен в blacklist
func (tbs *TokenBlacklistService) RevokeToken(ctx context.Context, tokenString string) error {
	// Парсим токен для получения информации о времени истечения
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return fmt.Errorf("не удалось разобрать токен: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return fmt.Errorf("неверный формат claims")
	}

	// Получаем время истечения токена
	exp, ok := claims["exp"].(float64)
	if !ok {
		return fmt.Errorf("отсутствует время истечения токена")
	}

	expiresAt := time.Unix(int64(exp), 0)
	now := time.Now()

	// Если токен уже истек, не добавляем его в blacklist
	if expiresAt.Before(now) {
		logger.Info(ctx, "Токен уже истек, не добавляем в blacklist")
		return nil
	}

	// Создаем ключ токена (UUID из первых 16 байт SHA-256) для экономии места
	tokenKey := tbs.tokenKey(tokenString)

	// TTL = время до истечения токена
	ttl := expiresAt.Sub(now)

	// Создаем фейковую сессию для записи в blacklist
	blacklistEntry := &model.Session{
		UUID:      tokenKey,
		UserUUID:  uuid.New(), // не важно для blacklist
		CreatedAt: now,
		UpdatedAt: now,
		ExpiresAt: expiresAt,
	}

	// Добавляем токен в blacklist с TTL
	err = tbs.cacheRepo.Set(ctx, tokenKey, blacklistEntry, ttl)
	if err != nil {
		logger.Error(ctx, "Ошибка добавления токена в blacklist", zap.Error(err))
		return fmt.Errorf("не удалось отозвать токен: %w", err)
	}

	logger.Info(ctx, "✅ Токен добавлен в blacklist", zap.String("token_key", tokenKey.String()))
	return nil
}

// IsTokenRevoked проверяет, находится ли токен в blacklist
func (tbs *TokenBlacklistService) IsTokenRevoked(ctx context.Context, tokenString string) bool {
	tokenKey := tbs.tokenKey(tokenString)

	// Пытаемся получить сессию из кеша по ключу токена
	session, err := tbs.cacheRepo.GetSessionByUUID(ctx, tokenKey)

	// Если токен найден в blacklist - он отозван
	return err == nil && session != nil
}

// RevokeAllUserTokens отзывает все токены пользователя
func (tbs *TokenBlacklistService) RevokeAllUserTokens(ctx context.Context, userUUID uuid.UUID) error {
	// Создаем UUID из строки для хранения в кеше
	revocationUUID := uuid.New()

	// Создаем фейковую сессию для записи отзыва всех токенов пользователя
	revocationEntry := &model.Session{
		UUID:      revocationUUID,
		UserUUID:  userUUID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		ExpiresAt: time.Now().Add(48 * time.Hour), // время жизни самого долгого refresh токена
	}

	// Устанавливаем маркер на 48 часов (время жизни самого долгого refresh токена)
	err := tbs.cacheRepo.Set(ctx, revocationUUID, revocationEntry, 48*time.Hour)
	if err != nil {
		logger.Error(ctx, "Ошибка отзыва всех токенов пользователя",
			zap.String("user_uuid", userUUID.String()),
			zap.Error(err))
		return fmt.Errorf("не удалось отозвать токены пользователя: %w", err)
	}

	logger.Info(ctx, "✅ Все токены пользователя отозваны", zap.String("user_uuid", userUUID.String()))
	return nil
}

// IsUserTokensRevoked проверяет, отозваны ли все токены пользователя
// Упрощенная версия - для будущих улучшений
func (tbs *TokenBlacklistService) IsUserTokensRevoked(ctx context.Context, userUUID uuid.UUID, tokenIssuedAt time.Time) bool {
	// Пока возвращаем false - можно реализовать позднее, когда будет нужно
	return false
}

// tokenKey создает детерминированный UUID из SHA256 хэша токена для использования в качестве ключа кеша.
// SHA256 дает 32 байта, а UUID вмещает 16, поэтому берем первые 16 байт.
func (tbs *TokenBlacklistService) tokenKey(tokenString string) uuid.UUID {
	hash := sha256.Sum256([]byte(tokenString))
	var key uuid.UUID
	copy(key[:], hash[:len(key)])
	return key
}
