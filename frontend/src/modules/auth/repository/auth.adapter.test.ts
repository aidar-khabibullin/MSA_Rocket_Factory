import { profileFromDto, registeredUserFromDto, registerRequestToDto, sessionFromDto } from './auth.adapter'

describe('auth adapters', () => {
  it('регистрация: плоские поля формы превращаются во вложенное тело', () => {
    expect(registerRequestToDto('demo', 'd@x.io', 'Secret123!')).toEqual({
      info: { user_info: { login: 'demo', email: 'd@x.io' }, password: 'Secret123!' },
    })
  })

  it('ответ регистрации и входа приводится к camelCase', () => {
    expect(registeredUserFromDto({ user_uuid: 'u1' })).toEqual({ userUuid: 'u1' })
    expect(sessionFromDto({ session_uuid: 's1' })).toEqual({ sessionUuid: 's1' })
  })

  it('me: профиль берётся из вложенного user, лишние поля сессии игнорируются', () => {
    const dto = {
      session: { uuid: 's1', created_at: '2026-11-08T12:00:00Z' },
      user: { uuid: 'u1', info: { login: 'demo', email: 'd@x.io', notification_methods: [] } },
    }

    expect(profileFromDto(dto)).toEqual({ userUuid: 'u1', login: 'demo', email: 'd@x.io' })
  })
})
