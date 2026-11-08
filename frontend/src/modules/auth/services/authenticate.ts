import { authApi } from '../repository/auth.api'
import type { Session } from '../store/session.store'

export type AuthMode = 'login' | 'register'

export interface Credentials {
  login: string
  email: string
  password: string
}

/** Регистрация (при необходимости), вход и получение профиля: три шага, зависящих друг от друга. */
export async function authenticate(mode: AuthMode, credentials: Credentials): Promise<Session> {
  const { login, email, password } = credentials

  if (mode === 'register') await authApi.register(login, email, password)
  const { sessionUuid } = await authApi.login(login, password)
  const me = await authApi.me(sessionUuid)

  return { sessionUuid, userUuid: me.userUuid, login: me.login }
}
