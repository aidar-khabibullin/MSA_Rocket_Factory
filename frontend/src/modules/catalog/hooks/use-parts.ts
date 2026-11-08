import { useMemo } from 'react'
import { useQuery } from '@tanstack/react-query'

import { useRequiredSession } from '@/modules/auth'

import type { Part } from '../model/part'
import { catalogApi } from '../repository/catalog.api'
import { catalogKeys } from '../repository/catalog.keys'

/** Каталог меняется редко: один запрос на всё приложение, общий для каталога, заказов и состава заказа. */
export function useParts() {
  const { sessionUuid } = useRequiredSession()
  return useQuery({
    queryKey: catalogKeys.parts(sessionUuid),
    queryFn: () => catalogApi.parts(sessionUuid),
    refetchOnWindowFocus: false,
    staleTime: 5 * 60_000,
  })
}

/** Индекс по UUID: детали заказов ищутся за O(1), а не перебором (правило js-index-maps). */
export function usePartsIndex(): ReadonlyMap<string, Part> | undefined {
  const { data } = useParts()
  return useMemo(() => (data ? new Map(data.map((part) => [part.uuid, part])) : undefined), [data])
}
