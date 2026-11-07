import { useMemo } from 'react'
import useSWR from 'swr'
import { create } from 'zustand'

import { catalogApi } from '@/shared/api/endpoints'
import { swrKeys } from '@/shared/api/keys'
import type { Part } from '@/shared/api/types'
import { useRequiredSession, useSessionStore } from '@/features/auth/session'

/** Каталог меняется редко: один запрос на всё приложение, общий для каталога, заказов и состава заказа. */
export function useParts() {
  const { sessionUuid } = useRequiredSession()
  return useSWR(swrKeys.parts(sessionUuid), ([, session]) => catalogApi.parts(session), {
    revalidateOnFocus: false,
    dedupingInterval: 5 * 60_000,
  })
}

/** Индекс по UUID: детали заказов ищутся за O(1), а не перебором (правило js-index-maps). */
export function usePartsIndex(): ReadonlyMap<string, Part> | undefined {
  const { data } = useParts()
  return useMemo(() => (data ? new Map(data.map((part) => [part.part_uuid, part])) : undefined), [data])
}

/**
 * Корзина — клиентское состояние без персистентности. Zustand-селекторы подписывают компонент
 * только на нужный кусок, а действия стабильны (не пересоздаются), поэтому memo-карточки не перерисовываются зря.
 */
interface CartState {
  selected: ReadonlySet<string>
  toggle: (partUuid: string) => void
  clear: () => void
}

export const useCartStore = create<CartState>()((set) => ({
  selected: new Set<string>(),
  toggle: (partUuid) =>
    set((state) => {
      const next = new Set(state.selected)
      if (!next.delete(partUuid)) next.add(partUuid)
      return { selected: next }
    }),
  clear: () => set({ selected: new Set<string>() }),
}))

// Корзина принадлежит сессии: после выхода она не должна достаться следующему пользователю.
useSessionStore.subscribe((state, previous) => {
  if (previous.session && !state.session) useCartStore.getState().clear()
})
