import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'

import { AuthPage } from './auth-page'

describe('AuthPage', () => {
  it('переключатель языка меняет тексты страницы и запоминает выбор', async () => {
    render(
      <MemoryRouter>
        <AuthPage />
      </MemoryRouter>,
    )
    expect(screen.getByRole('heading', { name: 'С возвращением' })).toBeInTheDocument()

    await userEvent.click(screen.getByRole('button', { name: /EN/ }))

    expect(screen.getByRole('heading', { name: 'Welcome back' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Sign in' })).toBeInTheDocument()
    expect(localStorage.getItem('rf:language:v1')).toContain('"language":"en"')
  })
})
