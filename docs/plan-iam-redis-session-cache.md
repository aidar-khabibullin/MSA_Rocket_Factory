# Plan: Redis — кеш сессий и чёрный список токенов IAM

**PRD:** [prd-iam-redis-session-cache.md](prd-iam-redis-session-cache.md)
**Дата:** 2026-11-08

## Фазы реализации

### Фаза 1: Кеш сессий (Tracer Bullet) — выполнена

**Цель:** Проверка сессии без обращения к PostgreSQL.
**Затрагивает:** backend
**Задачи:**

- [x] Redis-репозиторий сессий с TTL
- [x] Composite-репозиторий Redis + PostgreSQL
- [x] `CacheManager`: `GetSession`, `SetSession`, `DeleteSession`
- [x] Unit-тесты менеджера

**Когда готова:** повторный `GetSession` не идёт в PostgreSQL.

### Фаза 2: Чёрный список токенов — выполнена

**Цель:** Мгновенный отзыв токенов.
**Затрагивает:** backend
**Задачи:**

- [x] `TokenBlacklistService`: отзыв токена и всех токенов пользователя
- [x] `IsTokenRevoked`, `IsUserTokensRevoked`
- [x] Unit-тесты

**Когда готова:** токены, выпущенные до `RevokeAllUserTokens`, отклоняются.

### Фаза 3: Статистика и прогрев — выполнена

**Цель:** Видеть эффективность кеша.
**Затрагивает:** backend
**Задачи:**

- [x] Статистика hit/miss, `GetHitRatio`, `LogStats`, `ResetStats`
- [x] `WarmupCache`
- [x] Интеграционный тест Redis-репозитория сессий на реальном Redis (testcontainers): чтение, TTL, удаление, истечение. Composite-репозиторий (Redis + PostgreSQL) на реальных базах не проверялся

**Когда готова:** интеграционный тест проходит: `go test -tags=integration ./iam/internal/repository/redis/`.
