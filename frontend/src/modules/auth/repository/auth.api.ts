import { http } from '@/common/api/http'

/** Ответы iam в формате бэкенда (snake_case) приводятся к camelCase здесь, до выхода из репозитория. */
export const authApi = {
  register: async (login: string, email: string, password: string): Promise<{ userUuid: string }> => {
    const { user_uuid } = await http<{ user_uuid: string }>('/api/auth/register', {
      method: 'POST',
      body: { login, email, password },
    })
    return { userUuid: user_uuid }
  },

  login: async (login: string, password: string): Promise<{ sessionUuid: string }> => {
    const { session_uuid } = await http<{ session_uuid: string }>('/api/auth/login', {
      method: 'POST',
      body: { login, password },
    })
    return { sessionUuid: session_uuid }
  },

  me: async (sessionUuid: string): Promise<{ userUuid: string; login: string; email: string }> => {
    const { user_uuid, login, email } = await http<{ user_uuid: string; login: string; email: string }>(
      '/api/auth/me',
      { sessionUuid },
    )
    return { userUuid: user_uuid, login, email }
  },
}
