import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { act, renderHook, waitFor } from '@testing-library/react'
import type { ReactNode } from 'react'
import { MemoryRouter, useLocation, useNavigate } from 'react-router-dom'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { useSessionStore } from '@/modules/auth'

import { useCatalogPage } from './use-catalog-page'

const part = (uuid: string, name: string, category: string) => ({
  part_uuid: uuid,
  name,
  description: '',
  price: 100,
  stock_quantity: '5', // int64 в proto-JSON приходит строкой
  category,
  manufacturer: { name: 'ACME', country: 'US', url: '' },
  tags: [],
})

const PARTS = [
  part('p1', 'Двигатель Raptor', 'CATEGORY_ENGINE'),
  part('p2', 'Бак топливный', 'CATEGORY_FUEL'),
]

const setup = (initial: string) => {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  const wrapper = ({ children }: { children: ReactNode }) => (
    <QueryClientProvider client={client}>
      <MemoryRouter initialEntries={[initial]}>{children}</MemoryRouter>
    </QueryClientProvider>
  )
  return renderHook(() => ({ page: useCatalogPage(), location: useLocation(), navigate: useNavigate() }), { wrapper })
}

describe('useCatalogPage: фильтры в URL', () => {
  beforeEach(() => {
    useSessionStore.setState({ session: { sessionUuid: 's1', userUuid: 'u1', login: 'demo' } })
    vi.stubGlobal(
      'fetch',
      vi.fn(async () => new Response(JSON.stringify({ parts: PARTS }), { status: 200 })),
    )
  })
  afterEach(() => vi.unstubAllGlobals())

  it('читает поиск и категорию из адреса', async () => {
    const { result } = setup('/?q=двигатель&category=engine')

    await waitFor(() => expect(result.current.page.parts).toBeDefined())

    expect(result.current.page.query).toBe('двигатель')
    expect(result.current.page.category).toBe('CATEGORY_ENGINE')
    expect(result.current.page.visible.map((item) => item.uuid)).toEqual(['p1'])
  })

  it('игнорирует неизвестную категорию в адресе', async () => {
    const { result } = setup('/?category=nope')

    await waitFor(() => expect(result.current.page.parts).toBeDefined())

    expect(result.current.page.category).toBeNull()
    expect(result.current.page.visible).toHaveLength(2)
  })

  it('выбор категории записывается в адрес сразу', async () => {
    const { result } = setup('/')
    await waitFor(() => expect(result.current.page.parts).toBeDefined())

    act(() => result.current.page.setCategory('CATEGORY_FUEL'))

    expect(new URLSearchParams(result.current.location.search).get('category')).toBe('fuel')
    expect(result.current.page.visible.map((item) => item.uuid)).toEqual(['p2'])
  })

  it('поиск попадает в адрес после паузы, а не на каждую букву', async () => {
    vi.useFakeTimers({ shouldAdvanceTime: true })
    try {
      const { result } = setup('/')
      await waitFor(() => expect(result.current.page.parts).toBeDefined())

      act(() => result.current.page.setQuery('бак'))
      expect(result.current.location.search).toBe('')

      await act(async () => {
        await vi.advanceTimersByTimeAsync(300)
      })
      expect(new URLSearchParams(result.current.location.search).get('q')).toBe('бак')
    } finally {
      vi.useRealTimers()
    }
  })

  it('сброс фильтров очищает адрес', async () => {
    vi.useFakeTimers({ shouldAdvanceTime: true })
    try {
      const { result } = setup('/?q=бак&category=fuel')
      await waitFor(() => expect(result.current.page.parts).toBeDefined())

      act(() => result.current.page.resetFilters())
      await act(async () => {
        await vi.advanceTimersByTimeAsync(300)
      })

      expect(result.current.location.search).toBe('')
      expect(result.current.page.query).toBe('')
      expect(result.current.page.category).toBeNull()
    } finally {
      vi.useRealTimers()
    }
  })

  it('внешнее изменение адреса обновляет поле поиска', async () => {
    const { result } = setup('/?q=бак')
    await waitFor(() => expect(result.current.page.parts).toBeDefined())
    expect(result.current.page.query).toBe('бак')

    // Ссылка «Каталог» в шапке ведёт на «/» без параметров.
    act(() => void result.current.navigate('/'))

    await waitFor(() => expect(result.current.page.query).toBe(''))
    expect(result.current.page.visible).toHaveLength(2)
  })
})
