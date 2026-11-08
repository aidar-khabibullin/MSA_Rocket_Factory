/** Тела запросов и ответов iam через Envoy (proto-JSON, snake_case). Дальше репозитория они не уходят. */

export interface RegisterRequestDto {
  info: {
    user_info: { login: string; email: string }
    password: string
  }
}

export interface RegisterResponseDto {
  user_uuid: string
}

export interface LoginRequestDto {
  login: string
  password: string
}

export interface LoginResponseDto {
  session_uuid: string
}

/** Ответ GET /api/auth/me: сессия и вложенный пользователь. */
export interface WhoamiResponseDto {
  user: {
    uuid: string
    info: { login: string; email: string }
  }
}
