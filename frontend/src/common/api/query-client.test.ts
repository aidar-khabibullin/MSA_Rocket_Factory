import { afterEach, describe, expect, it, vi } from 'vitest'

import { ApiError } from './http'
import { onUnauthorized, queryClient } from './query-client'

describe('queryClient: обработка 401', () => {
  afterEach(() => {
    queryClient.clear()
    onUnauthorized(() => {})
  })

  it('вызывает обработчик при 401 в запросе', async () => {
    const handler = vi.fn<() => void>()
    onUnauthorized(handler)

    await queryClient
      .fetchQuery({ queryKey: ['x'], queryFn: () => Promise.reject(new ApiError(401, 'нет сессии')) })
      .catch(() => undefined)

    expect(handler).toHaveBeenCalledOnce()
  })

  it('вызывает обработчик при 401 в мутации', async () => {
    const handler = vi.fn<() => void>()
    onUnauthorized(handler)

    await queryClient
      .getMutationCache()
      .build(queryClient, { mutationFn: () => Promise.reject(new ApiError(401, 'нет сессии')) })
      .execute(undefined)
      .catch(() => undefined)

    expect(handler).toHaveBeenCalledOnce()
  })

  it('не реагирует на другие ошибки', async () => {
    const handler = vi.fn<() => void>()
    onUnauthorized(handler)

    await queryClient
      .fetchQuery({ queryKey: ['y'], queryFn: () => Promise.reject(new ApiError(500, 'сбой')) })
      .catch(() => undefined)

    expect(handler).not.toHaveBeenCalled()
  })
})
