import { http } from '@/common/api/http'

import type { Profile } from '../model/profile'
import { profileFromDto, registeredUserFromDto, registerRequestToDto, sessionFromDto } from './auth.adapter'
import type { LoginRequestDto, LoginResponseDto, RegisterResponseDto, WhoamiResponseDto } from './auth.dto'

export const authApi = {
  register: async (login: string, email: string, password: string): Promise<{ userUuid: string }> =>
    registeredUserFromDto(
      await http<RegisterResponseDto>('/api/auth/register', {
        method: 'POST',
        body: registerRequestToDto(login, email, password),
      }),
    ),

  login: async (login: string, password: string): Promise<{ sessionUuid: string }> =>
    sessionFromDto(
      await http<LoginResponseDto>('/api/auth/login', {
        method: 'POST',
        body: { login, password } satisfies LoginRequestDto,
      }),
    ),

  me: async (sessionUuid: string): Promise<Profile> =>
    profileFromDto(await http<WhoamiResponseDto>('/api/auth/me', { sessionUuid })),
}
