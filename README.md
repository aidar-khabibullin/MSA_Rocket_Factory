# MSA_Rocket_Factory

![Coverage](https://img.shields.io/endpoint?url=https://aidar-khabibullin.github.io/MSA_Rocket_Factory/coverage-badge.json)
[![CI](https://github.com/aidar-khabibullin/MSA_Rocket_Factory/actions/workflows/ci.yml/badge.svg)](https://github.com/aidar-khabibullin/MSA_Rocket_Factory/actions/workflows/ci.yml)

📊 **Отчеты о качестве кода:** 
- 📊 [Coverage Report](https://aidar-khabibullin.github.io/MSA_Rocket_Factory/) - покрытие бизнес-логики тестами (internal пакеты)
- 🧪 [Test Report](https://aidar-khabibullin.github.io/MSA_Rocket_Factory/tests/) - результаты unit и интеграционных тестов
- 🔍 [Lint Report](https://aidar-khabibullin.github.io/MSA_Rocket_Factory/lint/) - результаты статического анализа кода

> **Примечание:** Покрытие кода рассчитывается только для бизнес-логики (пакеты `internal/`), исключая `mocks`, `test` пакеты и `main.go` файлы.

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

Описание метрик и дашбордов: [docs/metrics.md](docs/metrics.md).

Архитектурные решения и известные ограничения: [docs/architecture.md](docs/architecture.md).

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

GitHub Actions (`.github/workflows`): линтер, unit- и интеграционные тесты и отчёт о покрытии (публикуется на GitHub Pages).
Версии Go и инструментов берутся из `Taskfile.yml`. Зависимости обновляет Dependabot (`.github/dependabot.yml`).
