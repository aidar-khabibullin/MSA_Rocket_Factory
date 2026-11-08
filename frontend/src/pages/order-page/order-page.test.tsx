import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'

import { useSessionStore } from '@/modules/auth'

import { OrderPage } from './order-page'

const order = (status: string) => ({
  order_uuid: 'o1',
  user_uuid: 'u1',
  part_uuids: ['p1'],
  total_price: 100,
  status,
})

const json = (body: unknown) => new Response(JSON.stringify(body), { status: 200 })

function renderPage() {
  // Свежий клиент на каждый тест: иначе кэш делится данными между тестами.
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  render(
    <QueryClientProvider client={client}>
      <MemoryRouter initialEntries={['/orders/o1']}>
        <Routes>
          <Route path="/orders/:orderId" element={<OrderPage />} />
        </Routes>
      </MemoryRouter>
    </QueryClientProvider>,
  )
}

describe('OrderPage: отмена заказа', () => {
  beforeEach(() => {
    useSessionStore.setState({ session: { sessionUuid: 's1', userUuid: 'u1', login: 'demo' } })
  })
  afterEach(() => vi.unstubAllGlobals())

  it('требует подтверждения и отменяет заказ только после него', async () => {
    let status = 'PENDING_PAYMENT'
    const fetchMock = vi.fn<(request: Request) => Promise<Response>>(async (request) => {
      const { pathname } = new URL(request.url)
      if (pathname.endsWith('/cancel')) {
        status = 'CANCELLED'
        return new Response(null, { status: 204 })
      }
      if (pathname.startsWith('/api/parts')) return json([])
      return request.method === 'POST' ? json({}) : json(order(status))
    })
    vi.stubGlobal('fetch', fetchMock)
    renderPage()

    await userEvent.click(await screen.findByRole('button', { name: 'Отменить заказ' }))

    // Один клик ничего не отправляет: сначала вопрос.
    const dialog = await screen.findByRole('alertdialog', { name: 'Отменить заказ?' })
    expect(fetchMock.mock.calls.some(([request]) => new URL(request.url).pathname.endsWith('/cancel'))).toBe(false)

    await userEvent.click(within(dialog).getByRole('button', { name: 'Да, отменить' }))

    await waitFor(() => expect(fetchMock.mock.calls.some(([request]) => new URL(request.url).pathname.endsWith('/cancel'))).toBe(true))
    expect(await screen.findByText('Заказ отменён', { selector: 'span' })).toBeInTheDocument()
  })

  it('отказ в диалоге не отправляет запрос', async () => {
    const fetchMock = vi.fn<(request: Request) => Promise<Response>>(async (request) =>
      new URL(request.url).pathname.startsWith('/api/parts') ? json([]) : json(order('PENDING_PAYMENT')),
    )
    vi.stubGlobal('fetch', fetchMock)
    renderPage()

    await userEvent.click(await screen.findByRole('button', { name: 'Отменить заказ' }))
    await userEvent.click(await screen.findByRole('button', { name: 'Оставить заказ' }))

    await waitFor(() => expect(screen.queryByRole('alertdialog')).not.toBeInTheDocument())
    expect(fetchMock.mock.calls.some(([request]) => new URL(request.url).pathname.endsWith('/cancel'))).toBe(false)
  })
})
