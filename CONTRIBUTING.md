# Участие в разработке

## Окружение

Нужны Go 1.24.4+, Docker с Compose, [Task](https://taskfile.dev/installation/) и Node.js 22+.
Остальные инструменты (golangci-lint, buf, ogen, mockery) ставятся в `./bin` командами `task`.
Быстрый старт описан в [README](README.md).

## Рабочий процесс

1. Создайте ветку от `main`: `feat/...`, `fix/...`, `docs/...`, `refactor/...`.
2. Для новой фичи сначала опишите требования в `docs/prd-*.md`, затем план в `docs/plan-*.md`
   (скиллы `prd`, `plan-phase`, `issues` в `.claude/skills`).
3. Перед коммитом запустите проверки:

   ```bash
   task format
   task lint
   task test
   task frontend:check   # если менялся фронтенд
   ```

4. После изменения `.proto` или OpenAPI выполните `task gen` и закоммитьте сгенерированный код.
5. Откройте pull request в `main`. CI должен быть зелёным: линтер, тесты, фронтенд и проверки безопасности.

## Коммиты

Используем [Conventional Commits](https://www.conventionalcommits.org/): `feat(order): ...`,
`fix(iam): ...`, `docs: ...`, `chore: ...`. Одно логическое изменение в одном коммите.

## Код

- Бизнес-логика живёт в `internal/service`, доступ к данным в `internal/repository`; покрытие считается по ним.
- Моки генерирует mockery (`task mockery:gen`), руками их не правим.
- Новый gRPC-сервис обязан регистрировать health-сервис (`platform/pkg/grpc/health`).
- Секреты в репозиторий не попадают: только шаблоны в `deploy/env/*.template`.
