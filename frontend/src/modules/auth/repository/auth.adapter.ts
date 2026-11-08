import type { Profile } from '../model/profile'
import type { LoginResponseDto, RegisterRequestDto, RegisterResponseDto, WhoamiResponseDto } from './auth.dto'

export const registerRequestToDto = (login: string, email: string, password: string): RegisterRequestDto => ({
  info: { user_info: { login, email }, password },
})

export const registeredUserFromDto = (dto: RegisterResponseDto): { userUuid: string } => ({ userUuid: dto.user_uuid })

export const sessionFromDto = (dto: LoginResponseDto): { sessionUuid: string } => ({ sessionUuid: dto.session_uuid })

export const profileFromDto = (dto: WhoamiResponseDto): Profile => ({
  userUuid: dto.user.uuid,
  login: dto.user.info.login,
  email: dto.user.info.email,
})
