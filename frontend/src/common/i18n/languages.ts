export const LANGUAGES = ['ru', 'en'] as const
export type Language = (typeof LANGUAGES)[number]

export const DEFAULT_LANGUAGE: Language = 'ru'

/** BCP 47 тег для Intl: форматы чисел и времени следуют выбранному языку. */
export const INTL_LOCALE: Record<Language, string> = { ru: 'ru-RU', en: 'en-US' }

export const isLanguage = (value: unknown): value is Language =>
  typeof value === 'string' && (LANGUAGES as readonly string[]).includes(value)
