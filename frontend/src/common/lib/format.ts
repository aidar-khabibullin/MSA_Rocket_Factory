import { currentLanguage, INTL_LOCALE, type Language } from '@/common/i18n'

// Intl-форматтеры дороги в создании: кэшируем по языку, а язык берём в момент вызова.
const priceFormats = new Map<Language, Intl.NumberFormat>()
const timeFormats = new Map<Language, Intl.DateTimeFormat>()

function priceFormat(language: Language): Intl.NumberFormat {
  let format = priceFormats.get(language)
  if (!format) {
    format = new Intl.NumberFormat(INTL_LOCALE[language], { minimumFractionDigits: 2, maximumFractionDigits: 2 })
    priceFormats.set(language, format)
  }
  return format
}

function timeFormat(language: Language): Intl.DateTimeFormat {
  let format = timeFormats.get(language)
  if (!format) {
    format = new Intl.DateTimeFormat(INTL_LOCALE[language], { hour: '2-digit', minute: '2-digit', second: '2-digit' })
    timeFormats.set(language, format)
  }
  return format
}

export const formatPrice = (value: number): string => priceFormat(currentLanguage()).format(value)
export const formatTime = (date: Date): string => timeFormat(currentLanguage()).format(date)
