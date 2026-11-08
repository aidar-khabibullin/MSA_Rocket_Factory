// Публичный API i18n: инициализация, регистрация словарей модулей и выбор языка.
export { currentLanguage, i18n, registerLocales } from './i18n'
export { useLanguageStore } from './language.store'
export { useRevalidateOnLanguageChange } from './use-revalidate-on-language-change'
export { DEFAULT_LANGUAGE, INTL_LOCALE, isLanguage, LANGUAGES, type Language } from './languages'
