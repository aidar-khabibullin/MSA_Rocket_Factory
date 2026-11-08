# PRD: Kafka — общая платформа сообщений

**Дата**: 2026-11-08
**Статус**: Implemented

## Цель

Дать всем сервисам единые producer, consumer и контракты событий, чтобы обмен `order.paid` и `ship.assembled` был одинаковым и наблюдаемым.

## Пользовательские сценарии

- Сервис публикует событие -> все заинтересованные группы получают его независимо друг от друга.
- Сервис потребляет события -> обработчик получает типизированное сообщение, ошибка не роняет процесс.
- Разработчик добавляет нового потребителя -> подключает его со своей consumer group, не меняя producer.
- Оператор смотрит метрики Kafka -> видит отправленные и обработанные сообщения.

## В скоупе

- Пакет `platform/pkg/kafka`: интерфейсы, `Message`, producer (`sarama.SyncProducer`), consumer group с цепочками middleware.
- События в protobuf (`shared/proto/events/v1`): `order.paid`, `ship.assembled`, у каждого есть `event_uuid`.
- Конфигурация producer: `acks=all`, 3 повтора отправки.
- Consumer отмечает offset (`MarkMessage`) только после успешной обработки (at-least-once).
- Middleware: трассировка, метрики Kafka, логирование.
- Топики и группы: `assembly-service`, `notification-service`, потребитель `ship.assembled` в `order`.

## Не в скоупе

- Схема-реестр и версионирование событий сверх protobuf `v1`.
- Транзакционная публикация Kafka (exactly-once).
- Надёжность доставки: повторы, DLQ, дедупликация, ключ партиции — [prd-reliable-kafka-event-delivery.md](prd-reliable-kafka-event-delivery.md).

## Технические ограничения

- Гарантия at-least-once: дубликаты возможны, обработчики должны быть идемпотентными.
- Ошибка обработчика сейчас не блокирует consumer, а следующее успешное сообщение сдвигает offset (`architecture.md`, п. 3).
- Ключ сообщения — случайный `event_uuid`, порядок событий одного заказа не гарантирован.

## Критерии готовности

- [x] Producer отправляет сообщение с ключом и значением и логирует topic, partition, offset.
- [x] Consumer вызывает обработчик для каждого сообщения и отмечает offset только при успехе.
- [x] Middleware tracing и metrics подключаются цепочкой.
- [x] События `order.paid` и `ship.assembled` описаны в protobuf и сгенерированы `task proto:gen`.
- [x] Метрики Kafka описаны в [metrics.md](metrics.md).
- [x] Повторы, DLQ, дедупликация и ключ `order_uuid` (отдельный PRD).
