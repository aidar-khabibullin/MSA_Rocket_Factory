import { mutate } from 'swr'
import { create } from 'zustand'
import { persist } from 'zustand/middleware'

import { persistStorage, STORAGE_KEYS } from '@/common/lib/storage'

export interface Session {
  sessionUuid: string
  userUuid: string
  login: string
}

interface SessionState {
  session: Session | null
  signIn: (session: Session) => void
  signOut: () => void
}

/** Данные из localStorage недоверенные: принимаем только объект ожидаемой формы. */
export function parseSession(raw: unknown): Session | null {
  if (typeof raw !== 'object' || raw === null) return null
  const { sessionUuid, userUuid, login } = raw as Record<string, unknown>
  if (typeof sessionUuid !== 'string' || typeof userUuid !== 'string' || typeof login !== 'string') return null
  return { sessionUuid, userUuid, login }
}

export const useSessionStore = create<SessionState>()(
  persist(
    (set) => ({
      session: null,
      signIn: (session) => set({ session }),
      signOut: () => {
        set({ session: null })
        // Кэш SWR принадлежит пользователю: после выхода его данные не должны остаться в памяти.
        void mutate(() => true, undefined, { revalidate: false })
      },
    }),
    {
      name: STORAGE_KEYS.session,
      version: 1,
      storage: persistStorage,
      partialize: ({ session }) => ({ session }),
      merge: (persisted, current) => ({
        ...current,
        session: parseSession((persisted as { session?: unknown } | undefined)?.session),
      }),
    },
  ),
)

// Вход или выход в другой вкладке: перечитываем хранилище (один слушатель на модуль).
window.addEventListener('storage', (event) => {
  if (event.key === STORAGE_KEYS.session) void useSessionStore.persist.rehydrate()
})

export const getSession = (): Session | null => useSessionStore.getState().session
export const signIn = (session: Session): void => useSessionStore.getState().signIn(session)
export const signOut = (): void => useSessionStore.getState().signOut()

export const useSession = (): Session | null => useSessionStore((state) => state.session)

/** Для страниц за авторизацией: роутер уже отправил гостя на /login, поэтому сессия есть всегда. */
export function useRequiredSession(): Session {
  const session = useSession()
  if (!session) throw new Error('useRequiredSession вызван без активной сессии')
  return session
}
