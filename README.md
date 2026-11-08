# MSA_Rocket_Factory

![Coverage](https://img.shields.io/endpoint?url=https://aidar-khabibullin.github.io/MSA_Rocket_Factory/coverage-badge.json)
[![CI](https://github.com/aidar-khabibullin/MSA_Rocket_Factory/actions/workflows/ci.yml/badge.svg)](https://github.com/aidar-khabibullin/MSA_Rocket_Factory/actions/workflows/ci.yml)

📊 **Отчеты о качестве кода:** 
- 📊 [Coverage Report](https://aidar-khabibullin.github.io/MSA_Rocket_Factory/) - покрытие бизнес-логики тестами (internal пакеты)
- 🧪 [Test Report](https://aidar-khabibullin.github.io/MSA_Rocket_Factory/tests/) - результаты unit и интеграционных тестов
- 🔍 [Lint Report](https://aidar-khabibullin.github.io/MSA_Rocket_Factory/lint/) - результаты статического анализа кода

> **Примечание:** Покрытие кода рассчитывается только для бизнес-логики (пакеты `internal/`), исключая `mocks`, `test` пакеты и `main.go` файлы.

## Быстрый старт

```bash
task up-all   # инфраструктура и все сервисы (первый запуск собирает образы, это несколько минут)
task web      # фронтенд на http://localhost:5173 (API через Envoy на :8085)
```

Дальше: зарегистрироваться на http://localhost:5173, выбрать детали, создать и оплатить заказ. Статус сменится
`PENDING_PAYMENT` → `PAID` → `ASSEMBLED`. Трассы смотрите в Jaeger, метрики в Grafana (адреса в таблице ниже).
Остановка: `task down-all`. Нужны Go, Docker, [Task](https://taskfile.dev/installation/) и Node.js (подробнее в разделе «Требования»).

## Описание

Проект: микросервисная система «Фабрика ракет» на Go (монорепозиторий с go.work).

![схема](docs/logo.png)

## Сервисы

| Сервис | Назначение | Транспорт | Хранилище |
|---|---|---|---|
| `envoy` | Единая точка входа для браузера: REST/JSON → gRPC (`iam`, `inventory`), прокси в `order`, CORS | HTTP (`:8085`) | - |
| `order` | Создание и оплата заказов | HTTP (OpenAPI, `:8080`), Kafka | PostgreSQL |
| `inventory` | Каталог деталей | gRPC (`:50051`) | MongoDB |
| `payment` | Оплата заказа | gRPC (`:50052`) | - |
| `iam` | Пользователи, сессии, JWT | gRPC (`:50053`) | PostgreSQL + Redis |
| `assembly` | Сборка ракеты после оплаты | Kafka | - |
| `notification` | Уведомления пользователю (Telegram) | Kafka, gRPC-клиент к `iam` | - |

Общий код: `platform/` (логгер, Kafka, middleware, трассировка, метрики) и `shared/` (protobuf/OpenAPI-контракты).

### Поток запроса из браузера

```
Браузер ──HTTP /api──▶ Envoy (:8085) ──gRPC──▶ iam        (/api/auth/register · login · me)
                           │         ──gRPC──▶ inventory  (/api/parts)
                           └─────────HTTP────▶ order      (/api/v1/orders...)
```

Envoy переводит REST/JSON в gRPC по аннотациям `google.api.http` в proto (gRPC-JSON transcoder),
передаёт сессию заголовком `Session-Uuid` и подставляет реальный адрес клиента в `x-forwarded-for`.
Конфигурация: `deploy/compose/envoy/envoy.yaml`, проверка: `task envoy:test` (нужен поднятый стек).
Dev-сервер фронтенда (`task web`) проксирует весь `/api` на Envoy. Подробности: [docs/architecture.md](docs/architecture.md).

### Поток заказа

```
Клиент ──HTTP──▶ order ──gRPC──▶ inventory   (подбор деталей)
                   │     ──gRPC──▶ payment     (оплата)
                   │     ──gRPC──▶ iam         (проверка пользователя)
                   │
                   └─ Kafka: order.paid ──▶ assembly ──▶ Kafka: ship.assembled ──┐
                                   │                                             │
                                   └──▶ notification (Telegram) ◀────────────────┤
                                                                  order ◀────────┘ (статус заказа)
```

Топики Kafka: `order.paid` (публикует `order`) и `ship.assembled` (публикует `assembly`).

## Требования

- [Go](https://go.dev/dl/) 1.24.4+
- [Docker](https://docs.docker.com/get-docker/) и Docker Compose
- [Task](https://taskfile.dev/installation/)
- Node.js (для `redocly` при сборке OpenAPI-спецификаций)

Остальные инструменты (golangci-lint, buf, ogen, mockery, envsubst и др.) устанавливаются в `./bin` командами `task`.

## Запуск приложения в Docker Compose

```bash
task up-all      # поднять core-инфраструктуру и все сервисы
task down-all    # остановить и удалить контейнеры
```

При первом запуске `.env`-файлы создаются автоматически из шаблонов (`task env:generate`):
значения по умолчанию берутся из `deploy/env/.env.template`, а `deploy/env/.env` и `deploy/compose/*/.env`
в git не хранятся. Перед использованием вне локальной машины замените пароли и JWT-секреты.

Отдельные сервисы: `task up-<service>` / `task down-<service>` (`inventory`, `payment`, `order`, `iam`, `assembly`, `notification`, `envoy`).

### Инфраструктура (core)

| Компонент | URL |
|---|---|
| Grafana | http://localhost:3000 |
| Prometheus | http://localhost:9090 |
| Jaeger | http://localhost:16686 |
| Kafka UI | http://localhost:8090 |
| Envoy (API для фронтенда) | http://localhost:8085 |
| Фронтенд (`task web`) | http://localhost:5173 |

Логин и пароль Grafana по умолчанию: `admin` / `administrator` (см. `deploy/env/.env.template`).

Описание метрик и дашбордов: [docs/metrics.md](docs/metrics.md).

Архитектурные решения и известные ограничения: [docs/architecture.md](docs/architecture.md). Все документы, PRD и планы фич: [docs/README.md](docs/README.md).

Нагрузочный тест и трассировка: [docs/load-testing.md](docs/load-testing.md).

## Разработка

Полный список команд: `task --list`.

```bash
task test            # unit-тесты всех модулей (с -race)
task test-coverage   # покрытие бизнес-логики (internal/service, internal/repository)
task lint            # golangci-lint
task format          # gofumpt + gci
task gen             # генерация кода (proto, OpenAPI)
task mockery:gen     # генерация моков
```

Интеграционные тесты помечены build-тегом `integration` и запускаются отдельно: `task test-integration`.
Часть тестов использует testcontainers, поэтому для них нужен запущенный Docker.

> **mockery и версия Go.** Mockery v2.53.x не умеет загружать пакеты под Go 1.25+ (ошибка
> `package "context" without types`). Если локально стоит более новый Go, запускайте генерацию так:
> `GOTOOLCHAIN=go1.24.4 task mockery:gen`.

## CI

GitHub Actions (`.github/workflows`):

- **Go:** golangci-lint, unit- и интеграционные тесты, отчёт о покрытии (публикуется на GitHub Pages).
- **Фронтенд:** typecheck, lint, тесты и сборка.
- **Безопасность:** `govulncheck` для всех Go-модулей, `gitleaks` (секреты), `npm audit` и Trivy для Dockerfile и compose.

Версии Go и инструментов берутся из `Taskfile.yml`.

## Надёжность и безопасность

- **Health-проверки.** gRPC-сервисы (`iam`, `inventory`, `payment`) отдают стандартный `grpc.health.v1`, `order` отдаёт `GET /health`;
  docker compose проверяет их через `healthcheck` (`docker ps` показывает `healthy`). У `assembly` и `notification` нет входящего
  порта: это Kafka-потребители, их состояние видно по метрикам и логам.
- **Envoy** ограничивает частоту запросов (429) и размер тела (413); перебор паролей ограничивает `iam`.
- **TLS** в compose не включён, схема для продакшена описана в [docs/architecture.md](docs/architecture.md#22-tls).
- Об уязвимостях сообщайте по [SECURITY.md](SECURITY.md).

## Участие и лицензия

Правила работы с репозиторием: [CONTRIBUTING.md](CONTRIBUTING.md). Лицензия: [MIT](LICENSE).

> Каталоги `.agents/` и `.claude/skills/` содержат скиллы для AI-ассистента и к коду проекта не относятся (см. [.agents/README.md](.agents/README.md)).
