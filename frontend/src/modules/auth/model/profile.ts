/** Владелец текущей сессии в терминах фронтенда (см. repository/auth.adapter.ts). */
export interface Profile {
  userUuid: string
  login: string
  email: string
}
