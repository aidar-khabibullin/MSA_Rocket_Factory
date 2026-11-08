import { create } from 'zustand'
import { persist } from 'zustand/middleware'

import { persistStorage, STORAGE_KEYS } from '@/common/lib/storage'

export type Theme = 'light' | 'dark'

interface ThemeState {
  theme: Theme
  setTheme: (theme: Theme) => void
  toggle: () => void
}

/** Начальную тему до отрисовки ставит inline-скрипт в index.html (учитывая сохранённую и системную). */
const domTheme = (): Theme => (document.documentElement.classList.contains('light') ? 'light' : 'dark')

function applyTheme(theme: Theme): void {
  const root = document.documentElement
  root.classList.remove('light', 'dark')
  root.classList.add(theme)
  root.setAttribute('data-theme', theme)
}

const isTheme = (value: unknown): value is Theme => value === 'light' || value === 'dark'

export const useThemeStore = create<ThemeState>()(
  persist(
    (set, get) => ({
      theme: domTheme(),
      setTheme: (theme) => set({ theme }),
      toggle: () => set({ theme: get().theme === 'dark' ? 'light' : 'dark' }),
    }),
    {
      name: STORAGE_KEYS.theme,
      version: 1,
      storage: persistStorage,
      partialize: ({ theme }) => ({ theme }),
      // Из хранилища принимаем только валидное значение.
      merge: (persisted, current) => {
        const theme = (persisted as Partial<ThemeState> | undefined)?.theme
        return { ...current, theme: isTheme(theme) ? theme : current.theme }
      },
    },
  ),
)

// DOM синхронизируется с состоянием в одном месте: при старте (после чтения хранилища) и при каждом изменении.
applyTheme(useThemeStore.getState().theme)
useThemeStore.subscribe(({ theme }) => applyTheme(theme))

export const useTheme = () => {
  const theme = useThemeStore((state) => state.theme)
  const toggle = useThemeStore((state) => state.toggle)
  return { theme, toggle }
}
