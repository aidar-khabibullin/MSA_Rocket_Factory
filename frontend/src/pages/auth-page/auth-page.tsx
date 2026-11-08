import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Navigate } from 'react-router-dom'

import { Logo } from '@/common/components/logo'
import { LanguageToggle } from '@/common/components/language-toggle'
import { ThemeToggle } from '@/common/components/theme-toggle'

import { AuthForm, useSession } from '@/modules/auth'
import type { AuthMode } from '@/modules/auth'

const YEAR = new Date().getFullYear()

const BENEFITS = ['catalog', 'flow', 'notify'] as const

export function AuthPage() {
  const { t } = useTranslation('auth')
  const session = useSession()
  const [mode, setMode] = useState<AuthMode>('login')

  if (session) return <Navigate to="/" replace />

  return (
    <div className="grid min-h-dvh lg:grid-cols-2">
      <aside className="relative hidden flex-col justify-between overflow-hidden border-r border-border/60 p-12 lg:flex">
        <Logo size="lg" />
        <div className="flex max-w-md flex-col gap-4">
          <h2 className="text-4xl font-bold leading-tight tracking-tight">
            {t('promo.headline')}
          </h2>
          <ul className="flex flex-col gap-2 text-muted">
            {BENEFITS.map((benefit) => (
              <li key={benefit}>{t(`promo.benefits.${benefit}`)}</li>
            ))}
          </ul>
        </div>
        <p className="text-sm text-muted">© {YEAR} Rocket Factory</p>
      </aside>

      <main className="relative flex items-center justify-center px-4 py-10">
        <div className="absolute right-4 top-4 flex items-center gap-1">
          <LanguageToggle />
          <ThemeToggle />
        </div>

        <div className="rise flex w-full max-w-sm flex-col gap-8">
          <div className="lg:hidden">
            <Logo />
          </div>
          <div>
            <h1 className="text-2xl font-bold tracking-tight">{t(`title.${mode}`)}</h1>
            <p className="mt-1 text-sm text-muted">{t(`subtitle.${mode}`)}</p>
          </div>

          <AuthForm key={mode} mode={mode} />

          <div className="flex flex-col items-center gap-1 border-t border-border/60 pt-6 text-sm">
            <span className="text-muted">{t(`switch.${mode}.prompt`)}</span>
            <button
              type="button"
              className="min-h-11 rounded-lg px-3 font-medium text-accent underline-offset-4 hover:underline"
              onClick={() => setMode(mode === 'login' ? 'register' : 'login')}
            >
              {t(`switch.${mode}.action`)}
            </button>
          </div>
        </div>
      </main>
    </div>
  )
}
