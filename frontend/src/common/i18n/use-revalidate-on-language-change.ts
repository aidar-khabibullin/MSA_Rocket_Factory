import { useEffect, useEffectEvent, useRef, type RefObject } from 'react'
import { useTranslation } from 'react-i18next'

/**
 * Тексты ошибок валидации вычисляются в момент проверки, поэтому уже показанная ошибка остаётся на прежнем
 * языке. При смене языка заново «показываем» ошибки только у полей, которые уже помечены невалидными:
 * нетронутые поля не краснеют, а фокус не переезжает (в отличие от form.reportValidity()).
 */
export function useRevalidateOnLanguageChange(formRef: RefObject<HTMLFormElement | null>): void {
  const { i18n } = useTranslation()
  const { language } = i18n
  const previousLanguage = useRef(language)

  const revalidate = useEffectEvent(() => {
    const invalid = formRef.current?.querySelectorAll('input[aria-invalid="true"]') ?? []
    // react-aria слушает нативное событие change и по нему фиксирует актуальный текст ошибки (при invalid
    // уже показанная ошибка не обновляется). Значение поля не меняется, поэтому React-обработчики не срабатывают.
    for (const input of invalid) input.dispatchEvent(new Event('change'))
  })

  useEffect(() => {
    if (previousLanguage.current === language) return // первый рендер и повторные срабатывания эффекта
    previousLanguage.current = language
    revalidate()
  }, [language])
}
