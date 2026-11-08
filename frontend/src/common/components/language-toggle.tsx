import { Button } from '@heroui/react/button'
import { useTranslation } from 'react-i18next'

import { LANGUAGES, useLanguageStore } from '@/common/i18n'

/** Кнопка показывает язык, на который переключит: так не нужен выпадающий список для двух вариантов. */
export const LanguageToggle = () => {
  const { t } = useTranslation()
  const language = useLanguageStore((state) => state.language)
  const setLanguage = useLanguageStore((state) => state.setLanguage)
  const next = LANGUAGES[(LANGUAGES.indexOf(language) + 1) % LANGUAGES.length] ?? language

  return (
    <Button
      size="sm"
      variant="ghost"
      onPress={() => setLanguage(next)}
      aria-label={t('language.switchTo', { language: next.toUpperCase() })}
    >
      {next.toUpperCase()}
    </Button>
  )
}
