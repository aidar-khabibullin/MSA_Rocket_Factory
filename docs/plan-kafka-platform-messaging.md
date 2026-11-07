# Plan: Kafka — общая платформа сообщений

**PRD:** [prd-kafka-platform-messaging.md](prd-kafka-platform-messaging.md)
**Дата:** 2026-11-08

## Фазы реализации

### Фаза 1: Producer, consumer и контракты (Tracer Bullet) — выполнена

**Цель:** `order` публикует `order.paid`, `assembly` его читает.
**Затрагивает:** backend, shared
**Задачи:**

- [x] События `order.paid` и `ship.assembled` в protobuf (`task proto:gen`)
- [x] Producer с `acks=all` и 3 повторами
- [x] Consumer group с отметкой offset только после успеха
- [x] Unit-тесты

**Когда готова:** сообщение проходит от `order` до `assembly`.

### Фаза 2: Middleware и метрики — выполнена

**Цель:** Наблюдаемость обмена.
**Затрагивает:** backend
**Задачи:**

- [x] Middleware трассировки и метрик Kafka
- [x] Описание метрик в [metrics.md](metrics.md)

**Когда готова:** метрики Kafka видны на дашборде.

### Фаза 3: Надёжность доставки

**Цель:** Повторы, DLQ, дедупликация, ключ партиции.
**Затрагивает:** backend
**Задачи:**

- [x] Реализуется в [plan-reliable-kafka-event-delivery.md](plan-reliable-kafka-event-delivery.md)

**Когда готова:** выполнен план надёжной доставки.
