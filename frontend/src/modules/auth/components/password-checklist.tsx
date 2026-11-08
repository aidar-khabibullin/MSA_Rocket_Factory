import Check from '@gravity-ui/icons/Check'
import { useTranslation } from 'react-i18next'

import { PASSWORD_RULES } from '../model/password-rules'

/** Чек-лист требований: статус передаётся иконкой и текстом для скринридера, а не только цветом. */
export const PasswordChecklist = ({ value }: { value: string }) => {
  const { t } = useTranslation('auth')
  return (
    <ul className="grid grid-cols-1 gap-x-4 gap-y-1.5 pt-1 text-xs sm:grid-cols-2" aria-label={t('passwordRules.title')}>
      {PASSWORD_RULES.map((rule) => {
        const done = rule.test(value)
        return (
          <li key={rule.id} className={`flex items-center gap-2 ${done ? 'text-success' : 'text-muted'}`}>
            <span
              className={`grid size-4 shrink-0 place-items-center rounded-full border transition-colors ${
                done ? 'border-success bg-success text-background' : 'border-border'
              }`}
              aria-hidden="true"
            >
              {done ? <Check className="size-3" /> : null}
            </span>
            <span>{t(rule.labelKey)}</span>
            <span className="sr-only">{done ? t('passwordRules.done') : t('passwordRules.todo')}</span>
          </li>
        )
      })}
    </ul>
  )
}
