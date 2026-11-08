import { render, screen } from '@testing-library/react'
import { createMemoryRouter, Outlet, RouterProvider } from 'react-router-dom'

import { RouteError } from './route-error'

function renderBroken(inline: boolean) {
  const Broken = () => {
    throw new Error('boom')
  }
  const router = createMemoryRouter([
    {
      path: '/',
      element: (
        <div>
          <nav>шапка</nav>
          <Outlet />
        </div>
      ),
      children: [{ index: true, Component: Broken, errorElement: <RouteError inline={inline} /> }],
    },
  ])
  render(<RouterProvider router={router} />)
}

describe('RouteError', () => {
  beforeEach(() => vi.spyOn(console, 'error').mockImplementation(() => undefined))

  it('показывает сообщение и кнопку «на главную»', () => {
    renderBroken(false)
    expect(screen.getByRole('alert')).toHaveTextContent('Не удалось открыть страницу')
    expect(screen.getByRole('button', { name: 'На главную' })).toBeInTheDocument()
  })

  it('встроенный вариант добавляет кнопку перезагрузки', () => {
    renderBroken(true)
    expect(screen.getByRole('button', { name: 'Обновить' })).toBeInTheDocument()
    // Сбой страницы не уносит раскладку: шапка остаётся на экране.
    expect(screen.getByText('шапка')).toBeInTheDocument()
  })
})
