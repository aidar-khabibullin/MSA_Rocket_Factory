import { create } from 'zustand'

import { useSessionStore } from '@/modules/auth'

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
