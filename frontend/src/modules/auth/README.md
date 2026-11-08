# auth

Вход, регистрация и сессия пользователя.

- `store/session.store.ts`: сессия в Zustand с `persist` (валидируется при чтении, синхронизируется между вкладками); `signOut` очищает кэш SWR.
- `repository/auth.api.ts`: запросы к `/api/auth/*`, ответы приводятся к camelCase.
- `services/authenticate.ts`: регистрация (если нужна), вход и получение профиля.
- `model/password-rules.ts`: правила пароля и логина (зеркало правил `iam`); ошибки возвращаются ключами словаря.
- `components/`: `AuthForm`, `PasswordChecklist`.

Публичный API: `AuthForm`, `authenticate`, `AuthMode`, `getSession`, `signIn`, `signOut`, `useSession`, `useRequiredSession`, `useSessionStore`.
