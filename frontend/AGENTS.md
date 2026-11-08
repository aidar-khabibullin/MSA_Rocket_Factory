# AGENTS.md: фронтенд Rocket Factory

## Стек и запуск
- React 19 + TypeScript (strict) + Vite + HeroUI v3 (Tailwind v4). Серверные данные: SWR, клиентское состояние: Zustand.
- Команды (из `frontend/`): `npm run dev`, `npm run typecheck`, `npm run lint` (oxlint), `npm test` (vitest), `npm run build`.
  Из корня репозитория: `task frontend:check` (типы, линтер, тесты), `task web` (BFF + фронтенд).
- Dev-сервер проксирует `/api/auth`, `/api/parts`, `/api/health` на BFF (`:8095`) и `/api/v1` на `order` (`:8080`), см. `vite.config.ts`.
- Базовая линия: `typecheck`, `lint --deny-warnings`, `test` и `build` проходят без ошибок. Не оставляйте предупреждений линтера.

## Архитектура (FEOD): это проверяет линтер
Слои и допустимые импорты: `app → pages | modules | common`, `pages → modules | common`, `modules → modules | common`, `common → common`.
`global/` содержит стили, настройку тестов и декларации типов.

- Из другого слоя модуль импортируется только через публичный API `@/modules/<имя>` (`index.ts`). Внутри модуля используйте относительные импорты.
- Модули не образуют циклов. Если экрану нужны два взаимозависимых модуля, он живёт в `pages/` (так устроен `catalog-page`).
- Если правило мешает, переносите логику вниз по слоям или расширяйте публичный API модуля. Не отключайте правило.
- Имена файлов и папок в `src/` только kebab-case (`unicorn/filename-case`).
- Структура модуля: `components/`, `hooks/` (SWR), `model/` (чистая логика и типы, без React и сети), `repository/` (API, DTO, адаптеры, ключи SWR), `services/`, `store/` (Zustand), `locales/`. Пустые папки не создавайте.

## Данные
- Формат бэкенда (snake_case) живёт только в `repository/`: `*.dto.ts` описывает ответ, `*.adapter.ts` приводит его к модели (camelCase), `*.api.ts` возвращает уже модель. UI и хуки DTO не видят.
- Ключи SWR лежат рядом с репозиторием модуля (`*.keys.ts`).
- Сессия: `modules/auth` (`useRequiredSession`, `signOut`). 401 обрабатывает `SWRConfig` в `app/app.tsx`; отдельную логику выхода не добавляйте.

## Локализация (i18n)
- `i18next` + `react-i18next`, языки `ru` (по умолчанию) и `en`. Словари вшиты в бандл, инициализация синхронная: `common/i18n`.
- Каждый модуль владеет своими словарями: `modules/<имя>/locales/{ru,en}.json`, регистрируются в `locales/index.ts` и подключаются из `index.ts` модуля. Общие строки: `common/i18n/locales/*/common.json`.
- В компонентах: `const { t } = useTranslation('<namespace>')`. Ключи типизированы по русским словарям (`global/i18next.d.ts`): опечатка ломает `typecheck`.
- Новый namespace: добавьте словарь, `registerLocales`, импорт в `index.ts` модуля, тип в `global/i18next.d.ts` и импорт в `global/test-setup.ts`.
- Любой ключ добавляется в оба языка сразу: `global/locales.test.ts` сравнивает наборы ключей и плейсхолдеры.
- В `model/` текст не хранится: модель возвращает ключи или коды, перевод делает UI. Даты и числа форматируйте через `common/lib/format.ts` (учитывает язык).
- Не вставляйте пользовательские строки в код напрямую.

## Страницы
- Страница в `pages/<имя>/` — разметка; вся логика (данные, фильтры, действия, состояние диалогов) живёт в facade-хуке рядом: `use-<имя>.ts`. Хук тестируется и читается отдельно от JSX.
- Необратимые действия подтверждаются через `ConfirmDialog` (`common/components`), он управляется снаружи (`isOpen`). Пример: отмена заказа.
- Сбой страницы или загрузки её чанка ловит `errorElement` вложенного маршрута (`RouteError inline` в `app/router.tsx`): шапка и навигация остаются. Общий `RouteError` на весь экран нужен только для сбоя раскладки и страницы входа.

## Уведомления и загрузка
- Всплывающие уведомления: `toastService` (`common/service/toast-service.ts`, поверх HeroUI Toast). Ошибки конкретной формы остаются inline через `ErrorAlert`.
- Ветвление «загрузка / нет данных / данные»: `AsyncGate` (`common/components`).

## Тесты
- Vitest + Testing Library; настройка в `global/test-setup.ts` (русский язык, in-memory `localStorage`).
- Чистую логику (`model/`, адаптеры) покрывайте юнит-тестами рядом с файлом.
