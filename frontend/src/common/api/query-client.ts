import { MutationCache, QueryCache, QueryClient } from '@tanstack/react-query'

import { ApiError } from './http'

let unauthorizedHandler: (() => void) | undefined

/**
 * Реакция на 401 регистрируется приложением (`app/`): слой `common` ничего не знает о модуле auth,
 * иначе получился бы цикл `common` → `auth` → `common`.
 */
export function onUnauthorized(handler: () => void): void {
  unauthorizedHandler = handler
}

function handleError(error: unknown): void {
  if (error instanceof ApiError && error.status === 401) unauthorizedHandler?.()
}

/** Единый клиент серверного состояния: 401 обрабатывается один раз и для запросов, и для мутаций. */
export const queryClient = new QueryClient({
  queryCache: new QueryCache({ onError: handleError }),
  mutationCache: new MutationCache({ onError: handleError }),
  defaultOptions: {
    queries: { retry: false },
  },
})
