import { create } from 'zustand'
import { persist } from 'zustand/middleware'

import { persistStorage, STORAGE_KEYS } from '@/common/lib/storage'

import { currentLanguage, i18n } from './i18n'
import { isLanguage, type Language } from './languages'

interface LanguageState {
  language: Language
  setLanguage: (language: Language) => void
}

/** Источник правды — i18next (он выбрал язык до рендера); стор лишь сохраняет выбор и даёт реактивность. */
export const useLanguageStore = create<LanguageState>()(
  persist(
    (set) => ({
      language: currentLanguage(),
      setLanguage: (language) => {
        void i18n.changeLanguage(language)
        set({ language })
      },
    }),
    {
      name: STORAGE_KEYS.language,
      version: 1,
      storage: persistStorage,
      partialize: ({ language }) => ({ language }),
      merge: (persisted, current) => {
        const language = (persisted as Partial<LanguageState> | undefined)?.language
        return { ...current, language: isLanguage(language) ? language : current.language }
      },
    },
  ),
)
