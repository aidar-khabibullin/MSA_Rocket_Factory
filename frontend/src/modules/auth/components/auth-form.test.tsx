import { act, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'

import { useLanguageStore } from '@/common/i18n'

import { getSession, useSessionStore } from '../store/session.store'
import { AuthForm } from './auth-form'

const json = (body: unknown, status = 200) => new Response(JSON.stringify(body), { status })

describe('AuthForm', () => {
  beforeEach(() => useSessionStore.setState({ session: null }))
  afterEach(() => vi.unstubAllGlobals())

  it('вход: три запроса по порядку, затем сессия сохраняется', async () => {
    const fetchMock = vi
      .fn<typeof fetch>()
      .mockResolvedValueOnce(json({ session_uuid: 'sess-1' }))
      .mockResolvedValueOnce(json({ user_uuid: 'user-1', login: 'demo', email: 'd@x.io' }))
    vi.stubGlobal('fetch', fetchMock)

    render(<AuthForm mode="login" />)
    await userEvent.type(screen.getByLabelText(/логин/i), 'demo')
    await userEvent.type(screen.getByLabelText(/пароль/i), 'Secret123!')
    await userEvent.click(screen.getByRole('button', { name: 'Войти' }))

    await waitFor(() => expect(getSession()).toEqual({ sessionUuid: 'sess-1', userUuid: 'user-1', login: 'demo' }))
    expect(fetchMock.mock.calls.map((call) => call[0])).toEqual(['/api/auth/login', '/api/auth/me'])
  })

  it('показывает ошибку сервера и не создаёт сессию', async () => {
    vi.stubGlobal('fetch', vi.fn<typeof fetch>().mockResolvedValue(json({ message: 'invalid login or password' }, 401)))

    render(<AuthForm mode="login" />)
    await userEvent.type(screen.getByLabelText(/логин/i), 'demo')
    await userEvent.type(screen.getByLabelText(/пароль/i), 'wrong')
    await userEvent.click(screen.getByRole('button', { name: 'Войти' }))

    expect(await screen.findByRole('alert')).toHaveTextContent('invalid login or password')
    expect(getSession()).toBeNull()
  })

  it('регистрация: чек-лист пароля отмечает выполненные требования по мере ввода', async () => {
    render(<AuthForm mode="register" />)
    const rules = () => screen.getByRole('list', { name: 'Требования к паролю' })

    expect(rules()).toHaveTextContent('Заглавная буква: не выполнено')

    await userEvent.type(screen.getByLabelText(/пароль/i), 'Abcd1')

    expect(rules()).toHaveTextContent('Заглавная буква: выполнено')
    expect(rules()).toHaveTextContent('Цифра: выполнено')
    expect(rules()).toHaveTextContent('Не менее 8 символов: не выполнено')
    expect(rules()).toHaveTextContent('Спецсимвол (! ? #): не выполнено')
  })

  it('при входе чек-лист пароля не показывается', () => {
    render(<AuthForm mode="login" />)

    expect(screen.queryByRole('list', { name: 'Требования к паролю' })).not.toBeInTheDocument()
  })

  it('показанная ошибка валидации переводится при смене языка', async () => {
    render(<AuthForm mode="register" />)

    await userEvent.type(screen.getByLabelText(/логин/i), 'ab')
    await userEvent.tab() // уход с поля показывает ошибку
    expect(await screen.findByText('Минимум 3 символа')).toBeInTheDocument()

    await act(async () => useLanguageStore.getState().setLanguage('en'))

    expect(await screen.findByText('At least 3 characters')).toBeInTheDocument()
    expect(screen.queryByText('Минимум 3 символа')).not.toBeInTheDocument()
  })

  it('смена языка не показывает ошибки на нетронутых полях', async () => {
    render(<AuthForm mode="register" />)

    await act(async () => useLanguageStore.getState().setLanguage('en'))

    expect(screen.queryByText('At least 3 characters')).not.toBeInTheDocument()
    expect(screen.queryByText('The password does not meet the requirements')).not.toBeInTheDocument()
  })
})
