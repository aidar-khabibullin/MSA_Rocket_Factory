import { act, renderHook } from '@testing-library/react'
import type { ReactNode } from 'react'
import { MemoryRouter, useLocation } from 'react-router-dom'
import { describe, expect, it } from 'vitest'

import { useUrlParam } from './use-url-param'

const wrapper = (initial: string) => {
  const Wrapper = ({ children }: { children: ReactNode }) => (
    <MemoryRouter initialEntries={[initial]}>{children}</MemoryRouter>
  )
  return Wrapper
}

const render = (initial: string) =>
  renderHook(() => ({ param: useUrlParam('q'), location: useLocation() }), { wrapper: wrapper(initial) })

describe('useUrlParam', () => {
  it('читает значение из адреса', () => {
    const { result } = render('/?q=двигатель')
    expect(result.current.param[0]).toBe('двигатель')
  })

  it('возвращает null, когда параметра нет', () => {
    const { result } = render('/')
    expect(result.current.param[0]).toBeNull()
  })

  it('записывает значение и сохраняет остальные параметры', () => {
    const { result } = render('/?category=engine')

    act(() => result.current.param[1]('топливо'))

    const params = new URLSearchParams(result.current.location.search)
    expect(params.get('q')).toBe('топливо')
    expect(params.get('category')).toBe('engine')
  })

  it('null удаляет параметр', () => {
    const { result } = render('/?q=x&category=engine')

    act(() => result.current.param[1](null))

    expect(result.current.location.search).toBe('?category=engine')
  })
})
