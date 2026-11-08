# orders

Заказы: создание, оплата, отмена, статусы.

- `repository/`: `/api/v1/orders*`; `OrderDto` → `Order` через `order.adapter.ts` (нулевой UUID и `UNKNOWN` становятся отсутствием значения).
- `hooks/use-orders.ts`: запросы и мутации TanStack Query; заказ опрашивается, пока идёт сборка.
- `store/order-history.store.ts`: в API нет списка заказов пользователя, поэтому UUID заказов хранятся локально.
- `model/`: типы, статусы и их тон, шаги прогресса, группировка деталей.
- `components/`: `OrderStepper`, `OrderComposition`, `PaymentMethodPicker`.

Зависит от `auth` и `catalog` (состав заказа берёт данные деталей из каталога).
