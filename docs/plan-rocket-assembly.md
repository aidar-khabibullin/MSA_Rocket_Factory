# Plan: Сборка ракеты (assembly)

**PRD:** [prd-rocket-assembly.md](prd-rocket-assembly.md)
**Дата:** 2026-11-08

## Фазы реализации

### Фаза 1: Сборка по событию (Tracer Bullet) — выполнена

**Цель:** Оплаченный заказ превращается в `ship.assembled`.
**Затрагивает:** backend
**Задачи:**

- [x] Consumer `order.paid`, декодер protobuf
- [x] Сервис сборки с задержкой 5–15 с и отменой по контексту
- [x] Producer `ship.assembled`
- [x] Unit-тесты

**Когда готова:** один `order.paid` даёт один `ship.assembled`.

### Фаза 2: Метрики и health — выполнена

**Цель:** Видеть сборку.
**Затрагивает:** backend
**Задачи:**

- [x] Метрики старта, завершения, ошибок и длительности
- [x] Health-check

**Когда готова:** метрики описаны в [metrics.md](metrics.md).

### Фаза 3: Защита от повторной сборки

**Цель:** Повтор `order.paid` не собирает ракету дважды.
**Затрагивает:** backend
**Задачи:**

- [x] Дедупликация по `event_uuid` — реализуется в [plan-reliable-kafka-event-delivery.md](plan-reliable-kafka-event-delivery.md), фаза 4

**Когда готова:** второй `order.paid` с тем же `event_uuid` не публикует второй `ship.assembled`.
