# catalog

Каталог деталей и корзина.

- `repository/`: `GET /api/parts`; `PartDto` → `Part` через `part.adapter.ts`; ключи TanStack Query.
- `hooks/use-parts.ts`: `useParts` (кэш на всё приложение) и `usePartsIndex` (поиск по UUID за O(1)).
- `store/cart.store.ts`: корзина в Zustand (в памяти), очищается при выходе.
- `model/`: категории (и их значения в URL), фильтрация, уровни остатка.
- `components/`: `PartCard`, `CategoryFilter`, `CartBar`.

Страница каталога лежит в `pages/catalog-page`: ей нужны и каталог, и заказы, а модули не должны зависеть друг от друга по кругу.

Фильтры каталога хранятся в адресной строке (`?q=…&category=engine`): `common/lib/use-url-param.ts` пишет их через `replace`, поиск попадает в URL с задержкой 300 мс (`use-debounced-value.ts`).
