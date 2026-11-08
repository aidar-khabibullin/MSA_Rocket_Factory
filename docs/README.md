# Документация

## Для разработчика

| Документ | О чём |
|---|---|
| [architecture.md](architecture.md) | Почему система устроена так: Kafka, gRPC/HTTP, Envoy, повторы, TLS, известные ограничения |
| [metrics.md](metrics.md) | Метрики сервисов и дашборды Grafana |
| [load-testing.md](load-testing.md) | Нагрузочный тест и трассировка в Jaeger |

## Как развивается проект

Каждая фича проходит цепочку **PRD → план по фазам → issues**. Файлы лежат здесь же, пары `prd-*.md` и `plan-*.md` с одним суффиксом.

| Фича | PRD | План |
|---|---|---|
| Каталог деталей (inventory) | [prd](prd-inventory-catalog.md) | [plan](plan-inventory-catalog.md) |
| Оплата заказа (payment) | [prd](prd-payment-processing.md) | [plan](plan-payment-processing.md) |
| Управление заказами (order) | [prd](prd-order-management.md) | [plan](plan-order-management.md) |
| Аутентификация (iam) | [prd](prd-iam-authentication.md) | [plan](plan-iam-authentication.md) |
| Кэш сессий в Redis | [prd](prd-iam-redis-session-cache.md) | [plan](plan-iam-redis-session-cache.md) |
| Обмен сообщениями через Kafka | [prd](prd-kafka-platform-messaging.md) | [plan](plan-kafka-platform-messaging.md) |
| Надёжная доставка событий | [prd](prd-reliable-kafka-event-delivery.md) | [plan](plan-reliable-kafka-event-delivery.md) |
| Сборка ракеты (assembly) | [prd](prd-rocket-assembly.md) | [plan](plan-rocket-assembly.md) |
| Уведомления в Telegram | [prd](prd-telegram-notifications.md) | [plan](plan-telegram-notifications.md) |
| Envoy вместо самописного BFF | [prd](prd-replace-bff-with-envoy.md) | [plan](plan-replace-bff-with-envoy.md) |

Скриншоты Grafana и Jaeger лежат в [img/](img).
