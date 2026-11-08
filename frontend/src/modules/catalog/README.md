# catalog

Каталог деталей и корзина.

- `repository/`: `GET /api/parts`; `PartDto` → `Part` через `part.adapter.ts`; ключи SWR.
- `hooks/use-parts.ts`: `useParts` (кэш на всё приложение) и `usePartsIndex` (поиск по UUID за O(1)).
- `store/cart.store.ts`: корзина в Zustand (в памяти), очищается при выходе.
- `model/`: категории, фильтрация, уровни остатка.
- `components/`: `PartCard`, `CategoryFilter`, `CartBar`.

Страница каталога лежит в `pages/catalog-page`: ей нужны и каталог, и заказы, а модули не должны зависеть друг от друга по кругу.
