import i18next, { use } from 'i18next'
import { initReactI18next } from 'react-i18next'

import { STORAGE_KEYS } from '@/common/lib/storage'

import { DEFAULT_LANGUAGE, isLanguage, LANGUAGES, type Language } from './languages'
import en from './locales/en/common.json'
import ru from './locales/ru/common.json'

/** Язык до первой отрисовки: сохранённый выбор, затем язык браузера, затем язык по умолчанию. */
function detectLanguage(): Language {
  try {
    // Формат записи Zustand persist: {state: {language}, version}.
    const raw = localStorage.getItem(STORAGE_KEYS.language)
    const stored: unknown = raw ? (JSON.parse(raw) as { state?: { language?: unknown } }).state?.language : undefined
    if (isLanguage(stored)) return stored
  } catch {
    /* хранилище недоступно */
  }
  const browser = globalThis.navigator?.languages?.map((tag) => tag.slice(0, 2)).find(isLanguage)
  return browser ?? DEFAULT_LANGUAGE
}

// Словари вшиты в бандл, поэтому init синхронный: первый рендер уже на нужном языке и без мигания.
void use(initReactI18next).init({
  resources: { ru: { common: ru }, en: { common: en } },
  lng: detectLanguage(),
  fallbackLng: DEFAULT_LANGUAGE,
  supportedLngs: LANGUAGES,
  defaultNS: 'common',
  ns: ['common'],
  interpolation: { escapeValue: false }, // React сам экранирует
  initAsync: false,
  react: { useSuspense: false },
})

document.documentElement.lang = i18next.language
i18next.on('languageChanged', (language) => {
  document.documentElement.lang = language
})

export const i18n = i18next

/** Модули подключают свои переводы сами: common ничего не знает о вышележащих слоях. */
export function registerLocales(namespace: string, bundles: Record<Language, object>): void {
  for (const language of LANGUAGES) {
    i18next.addResourceBundle(language, namespace, bundles[language], true, true)
  }
}

export const currentLanguage = (): Language => (isLanguage(i18next.language) ? i18next.language : DEFAULT_LANGUAGE)
