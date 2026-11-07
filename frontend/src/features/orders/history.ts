import { create } from 'zustand'
import { persist } from 'zustand/middleware'

import { persistStorage, STORAGE_KEYS } from '@/shared/lib/storage'

/**
 * В API нет списка заказов пользователя, поэтому их идентификаторы запоминаются локально.
 * Хранится только список UUID по пользователям (минимум данных), остальное загружается с сервера.
 */
interface OrderHistoryState {
  byUser: Record<string, string[]>
  remember: (userUuid: string, orderUuid: string) => void
}

const EMPTY: readonly string[] = []

export const useOrderHistoryStore = create<OrderHistoryState>()(
  persist(
    (set) => ({
      byUser: {},
      remember: (userUuid, orderUuid) =>
        set((state) => {
          const ids = state.byUser[userUuid] ?? EMPTY
          if (ids.includes(orderUuid)) return state
          return { byUser: { ...state.byUser, [userUuid]: [orderUuid, ...ids] } }
        }),
    }),
    {
      name: STORAGE_KEYS.orders,
      version: 1,
      storage: persistStorage,
      partialize: ({ byUser }) => ({ byUser }),
      merge: (persisted, current) => {
        const byUser = (persisted as { byUser?: unknown } | undefined)?.byUser
        return { ...current, byUser: isOrderMap(byUser) ? byUser : current.byUser }
      },
    },
  ),
)

function isOrderMap(value: unknown): value is Record<string, string[]> {
  return (
    typeof value === 'object' &&
    value !== null &&
    Object.values(value).every((ids) => Array.isArray(ids) && ids.every((id) => typeof id === 'string'))
  )
}

export const rememberOrder = (userUuid: string, orderUuid: string): void =>
  useOrderHistoryStore.getState().remember(userUuid, orderUuid)

/** Селектор возвращает ссылку из состояния (или общий EMPTY), поэтому лишних перерисовок нет. */
export const useOrderHistory = (userUuid: string): readonly string[] =>
  useOrderHistoryStore((state) => state.byUser[userUuid] ?? EMPTY)
