import { createJSONStorage, type StateStorage } from 'zustand/middleware'

/**
 * Хранилище для Zustand persist (правило client-localstorage-schema): ключи версионируются
 * (`rf:<имя>:v<версия>`), а ошибки localStorage (приватный режим, квота, запрет) подавляются:
 * приложение продолжает работать, состояние живёт только в памяти.
 */
const safeStorage: StateStorage = {
  getItem: (key) => {
    try {
      return localStorage.getItem(key)
    } catch {
      return null
    }
  },
  setItem: (key, value) => {
    try {
      localStorage.setItem(key, value)
    } catch {
      /* хранилище недоступно */
    }
  },
  removeItem: (key) => {
    try {
      localStorage.removeItem(key)
    } catch {
      /* ignore */
    }
  },
}

export const persistStorage = createJSONStorage(() => safeStorage)

export const STORAGE_KEYS = {
  session: 'rf:session:v1',
  theme: 'rf:theme:v1',
  orders: 'rf:orders:v1',
  language: 'rf:language:v1',
} as const

const wrap = (state: unknown) => JSON.stringify({ state, version: 1 })

/** Переносит данные первой версии приложения (ключи без префикса версии) в формат persist и удаляет старые. */
export function migrateLegacyStorage(): void {
  try {
    const byUser: Record<string, unknown> = {}

    // Ключи собираем заранее: во время обхода мы удаляем записи, и индексы сдвигались бы.
    const keys = Array.from({ length: localStorage.length }, (_, index) => localStorage.key(index))

    for (const key of keys) {
      if (key === null) continue
      const raw = localStorage.getItem(key)
      if (raw === null) continue

      if (key === 'rf.session' && localStorage.getItem(STORAGE_KEYS.session) === null) {
        localStorage.setItem(STORAGE_KEYS.session, wrap({ session: JSON.parse(raw) }))
      } else if (key === 'rf.theme' && localStorage.getItem(STORAGE_KEYS.theme) === null) {
        localStorage.setItem(STORAGE_KEYS.theme, wrap({ theme: raw }))
      } else if (key.startsWith('rf.orders.')) {
        byUser[key.slice('rf.orders.'.length)] = JSON.parse(raw)
      } else if (key !== 'rf.session' && key !== 'rf.theme') {
        continue
      }
      localStorage.removeItem(key)
    }

    if (Object.keys(byUser).length > 0 && localStorage.getItem(STORAGE_KEYS.orders) === null) {
      localStorage.setItem(STORAGE_KEYS.orders, wrap({ byUser }))
    }
  } catch {
    /* повреждённые данные старой версии просто игнорируем */
  }
}
