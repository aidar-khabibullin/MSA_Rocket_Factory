import { afterEach, describe, expect, it, vi } from 'vitest'

import { ApiError, http } from './http'
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

  it('401 от Envoy в формате {code, message, details} доходит до обработчика с сообщением сервера', async () => {
    const handler = vi.fn<() => void>()
    onUnauthorized(handler)
    vi.stubGlobal(
      'fetch',
      vi.fn<typeof fetch>().mockResolvedValue(
        new Response(JSON.stringify({ code: 16, message: 'session not found', details: [] }), {
          status: 401,
          headers: { 'Content-Type': 'application/json' },
        }),
      ),
    )

    const error = await queryClient
      .fetchQuery({ queryKey: ['envoy-401'], queryFn: () => http('/api/auth/me', { sessionUuid: 'dead' }) })
      .catch((caught: unknown) => caught)
    vi.unstubAllGlobals()

    expect(error).toMatchObject({ status: 401, message: 'session not found' })
    expect(handler).toHaveBeenCalledOnce()
  })
})
