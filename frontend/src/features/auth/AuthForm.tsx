import { Button } from '@heroui/react/button'
import { FieldError } from '@heroui/react/field-error'
import { Form } from '@heroui/react/form'
import { Input } from '@heroui/react/input'
import { Label } from '@heroui/react/label'
import { Spinner } from '@heroui/react/spinner'
import { TextField } from '@heroui/react/textfield'
import { useState, useTransition, type SyntheticEvent } from 'react'

import { errorMessage } from '@/shared/api/http'
import { ErrorAlert } from '@/shared/ui/ErrorAlert'

import { authenticate, type AuthMode } from './authenticate'
import { PasswordChecklist } from './PasswordChecklist'
import { loginError, passwordError } from './passwordRules'
import { signIn } from './session'

const FIELD = 'flex flex-col gap-2'

/** Форму пересоздаёт родитель через key={mode}: пароль и ошибка сбрасываются без эффектов. */
export function AuthForm({ mode }: { mode: AuthMode }) {
  const isRegister = mode === 'register'
  const [password, setPassword] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [isPending, startTransition] = useTransition()

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
        setError(errorMessage(cause, 'Не удалось выполнить вход. Попробуйте ещё раз.'))
      }
    })
  }

  return (
    <Form className="flex flex-col gap-6" onSubmit={handleSubmit}>
      <TextField
        className={FIELD}
        isRequired
        name="login"
        autoComplete="username"
        validate={isRegister ? loginError : undefined}
      >
        <Label>Логин</Label>
        <Input />
        <FieldError />
      </TextField>

      {isRegister ? (
        <TextField className={FIELD} isRequired name="email" type="email" autoComplete="email">
          <Label>Email</Label>
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
        validate={isRegister ? passwordError : undefined}
      >
        <Label>Пароль</Label>
        <Input />
        {isRegister ? <PasswordChecklist value={password} /> : null}
        <FieldError />
      </TextField>

      <ErrorAlert message={error} />

      <Button className="mt-2" type="submit" isDisabled={isPending} fullWidth size="lg">
        {isPending ? <Spinner size="sm" color="current" /> : null}
        {isRegister ? 'Зарегистрироваться' : 'Войти'}
      </Button>
    </Form>
  )
}
