# system

Состояние серверной части для пользователя и разработчика.

- `hooks/use-service-status.ts`: опрос `/api/health` раз в 30 секунд, один запрос на все компоненты.
- `components/`: `StatusBanner` (виден пользователю при проблемах), `Footer`, `ServiceStatusPanel` (только в dev-сборке).
