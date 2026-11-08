import ArrowRightFromSquare from '@gravity-ui/icons/ArrowRightFromSquare'
import { Button } from '@heroui/react/button'
import { useTranslation } from 'react-i18next'
import { Navigate, Outlet } from 'react-router-dom'

import { signOut, useSession } from '@/modules/auth'
import { Footer, StatusBanner } from '@/modules/system'
import { Logo } from '@/common/components/logo'
import { LanguageToggle } from '@/common/components/language-toggle'
import { ThemeToggle } from '@/common/components/theme-toggle'

import { Navigation } from './navigation'

export const Layout = () => {
  const { t } = useTranslation()
  const session = useSession()

  // Выход (в том числе из другой вкладки или по 401): страницы ниже требуют сессию, поэтому уходим на /login.
  if (!session) return <Navigate to="/login" replace />

  return (
    <div className="flex min-h-dvh flex-col">
      <header className="sticky top-0 z-20 border-b border-border/60 bg-background/70 backdrop-blur-md">
        <div className="mx-auto flex h-16 max-w-6xl items-center justify-between gap-4 px-4 sm:px-6">
          <div className="flex items-center gap-6">
            <Logo />
            <Navigation className="hidden gap-1 sm:flex" />
          </div>
          <div className="flex items-center gap-2 sm:gap-3">
            <LanguageToggle />
            <ThemeToggle />
            <span
              className="grid size-8 place-items-center rounded-full bg-surface text-sm font-bold ring-1 ring-border"
              aria-hidden="true"
            >
              {session.login.slice(0, 1).toUpperCase()}
            </span>
            <span className="hidden text-sm text-muted md:inline">{session.login}</span>
            <Button size="sm" variant="ghost" onPress={signOut} aria-label={t('account.signOutAria')}>
              <ArrowRightFromSquare />
              <span className="hidden sm:inline">{t('account.signOut')}</span>
            </Button>
          </div>
        </div>
        <Navigation className="mx-auto flex max-w-6xl gap-1 px-4 pb-2 sm:hidden" />
      </header>

      <StatusBanner />
      <main className="mx-auto w-full max-w-6xl flex-1 px-4 py-8 sm:px-6">
        <Outlet />
      </main>
      <Footer />
    </div>
  )
}
