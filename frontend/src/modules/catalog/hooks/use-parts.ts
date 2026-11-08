import { useMemo } from 'react'
import useSWR from 'swr'

import { useRequiredSession } from '@/modules/auth'

import type { Part } from '../model/part'
import { catalogApi } from '../repository/catalog.api'
import { catalogKeys } from '../repository/catalog.keys'

/** Каталог меняется редко: один запрос на всё приложение, общий для каталога, заказов и состава заказа. */
export function useParts() {
  const { sessionUuid } = useRequiredSession()
  return useSWR(catalogKeys.parts(sessionUuid), ([, session]) => catalogApi.parts(session), {
    revalidateOnFocus: false,
    dedupingInterval: 5 * 60_000,
  })
}

/** Индекс по UUID: детали заказов ищутся за O(1), а не перебором (правило js-index-maps). */
export function usePartsIndex(): ReadonlyMap<string, Part> | undefined {
  const { data } = useParts()
  return useMemo(() => (data ? new Map(data.map((part) => [part.uuid, part])) : undefined), [data])
}
