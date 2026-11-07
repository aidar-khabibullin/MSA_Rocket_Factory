import { useState } from 'react'
import { Navigate } from 'react-router-dom'

import { Logo } from '@/shared/ui/Logo'
import { ThemeToggle } from '@/shared/ui/ThemeToggle'

import { AuthForm } from './AuthForm'
import type { AuthMode } from './authenticate'
import { useSession } from './session'

const YEAR = new Date().getFullYear()

const COPY = {
  login: {
    title: 'С возвращением',
    subtitle: 'Войдите, чтобы заказать ракету',
    switchPrompt: 'Ещё нет аккаунта?',
    switchAction: 'Создать аккаунт',
  },
  register: {
    title: 'Создайте аккаунт',
    subtitle: 'Это займёт меньше минуты',
    switchPrompt: 'Уже зарегистрированы?',
    switchAction: 'Войти в аккаунт',
  },
} as const

const BENEFITS = [
  'Каталог двигателей, топлива и узлов',
  'Оплата и сборка в одном потоке',
  'Уведомления о статусе заказа в Telegram',
] as const

export function AuthPage() {
  const session = useSession()
  const [mode, setMode] = useState<AuthMode>('login')
  const copy = COPY[mode]

  if (session) return <Navigate to="/" replace />

  return (
    <div className="grid min-h-dvh lg:grid-cols-2">
      <aside className="relative hidden flex-col justify-between overflow-hidden border-r border-border/60 p-12 lg:flex">
        <Logo size="lg" />
        <div className="flex max-w-md flex-col gap-4">
          <h2 className="text-4xl font-bold leading-tight tracking-tight">
            От выбора деталей до готовой ракеты за&nbsp;несколько кликов
          </h2>
          <ul className="flex flex-col gap-2 text-muted">
            {BENEFITS.map((benefit) => (
              <li key={benefit}>{benefit}</li>
            ))}
          </ul>
        </div>
        <p className="text-sm text-muted">© {YEAR} Rocket Factory</p>
      </aside>

      <main className="relative flex items-center justify-center px-4 py-10">
        <div className="absolute right-4 top-4">
          <ThemeToggle />
        </div>

        <div className="rise flex w-full max-w-sm flex-col gap-8">
          <div className="lg:hidden">
            <Logo />
          </div>
          <div>
            <h1 className="text-2xl font-bold tracking-tight">{copy.title}</h1>
            <p className="mt-1 text-sm text-muted">{copy.subtitle}</p>
          </div>

          <AuthForm key={mode} mode={mode} />

          <div className="flex flex-col items-center gap-1 border-t border-border/60 pt-6 text-sm">
            <span className="text-muted">{copy.switchPrompt}</span>
            <button
              type="button"
              className="min-h-11 rounded-lg px-3 font-medium text-accent underline-offset-4 hover:underline"
              onClick={() => setMode(mode === 'login' ? 'register' : 'login')}
            >
              {copy.switchAction}
            </button>
          </div>
        </div>
      </main>
    </div>
  )
}
