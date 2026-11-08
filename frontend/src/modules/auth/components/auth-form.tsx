import { Button } from '@heroui/react/button'
import { FieldError } from '@heroui/react/field-error'
import { Form } from '@heroui/react/form'
import { Input } from '@heroui/react/input'
import { Label } from '@heroui/react/label'
import { Spinner } from '@heroui/react/spinner'
import { TextField } from '@heroui/react/textfield'
import { useRef, useState, useTransition, type SyntheticEvent } from 'react'
import { useTranslation } from 'react-i18next'

import { errorMessage } from '@/common/api/http'
import { ErrorAlert } from '@/common/components/error-alert'
import { useRevalidateOnLanguageChange } from '@/common/i18n'

import { authenticate, type AuthMode } from '../services/authenticate'
import { PasswordChecklist } from './password-checklist'
import { loginError, passwordError } from '../model/password-rules'
import { signIn } from '../store/session.store'

const FIELD = 'flex flex-col gap-2'

/** Форму пересоздаёт родитель через key={mode}: пароль и ошибка сбрасываются без эффектов. */
export function AuthForm({ mode }: { mode: AuthMode }) {
  const { t } = useTranslation('auth')
  const isRegister = mode === 'register'
  const formRef = useRef<HTMLFormElement>(null)
  useRevalidateOnLanguageChange(formRef)
  const [password, setPassword] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [isPending, startTransition] = useTransition()

  const translateError = (key: 'validation.loginMin' | 'validation.passwordInvalid' | null) =>
    key === null ? null : t(key)

  function handleSubmit(event: SyntheticEvent<HTMLFormElement>) {
    event.preventDefault()
    const data = new FormData(event.currentTarget)
    const credentials = {
      login: String(data.get('login') ?? '').trim(),
      email: String(data.get('email') ?? '').trim(),
      password,
    }

    setError(null)
    startTransition(async () => {
      try {
        signIn(await authenticate(mode, credentials))
      } catch (cause) {
        setError(errorMessage(cause, t('form.failed')))
      }
    })
  }

  return (
    <Form ref={formRef} className="flex flex-col gap-6" onSubmit={handleSubmit}>
      <TextField
        className={FIELD}
        isRequired
        name="login"
        autoComplete="username"
        validate={isRegister ? (value) => translateError(loginError(value)) : undefined}
      >
        <Label>{t('form.login')}</Label>
        <Input />
        <FieldError />
      </TextField>

      {isRegister ? (
        <TextField className={FIELD} isRequired name="email" type="email" autoComplete="email">
          <Label>{t('form.email')}</Label>
          <Input />
          <FieldError />
        </TextField>
      ) : null}

      <TextField
        className={FIELD}
        isRequired
        name="password"
        type="password"
        autoComplete={isRegister ? 'new-password' : 'current-password'}
        value={password}
        onChange={setPassword}
        validate={isRegister ? (value) => translateError(passwordError(value)) : undefined}
      >
        <Label>{t('form.password')}</Label>
        <Input />
        {isRegister ? <PasswordChecklist value={password} /> : null}
        <FieldError />
      </TextField>

      <ErrorAlert message={error} />

      <Button className="mt-2" type="submit" isDisabled={isPending} fullWidth size="lg">
        {isPending ? <Spinner size="sm" color="current" /> : null}
        {isRegister ? t('form.submitRegister') : t('form.submitLogin')}
      </Button>
    </Form>
  )
}
