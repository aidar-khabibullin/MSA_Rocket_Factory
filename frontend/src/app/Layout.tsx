import ArrowRightFromSquare from '@gravity-ui/icons/ArrowRightFromSquare'
import { Button } from '@heroui/react/button'
import { Navigate, NavLink, Outlet } from 'react-router-dom'

import { signOut, useSession } from '@/features/auth/session'
import { Footer } from '@/features/system/Footer'
import { StatusBanner } from '@/features/system/StatusBanner'
import { Logo } from '@/shared/ui/Logo'
import { ThemeToggle } from '@/shared/ui/ThemeToggle'

import { preloadPage, type PageName } from './routes'

const NAV: ReadonlyArray<{ to: string; label: string; page: PageName; end?: boolean }> = [
  { to: '/', label: 'Каталог', page: 'catalog', end: true },
  { to: '/orders', label: 'Мои заказы', page: 'orders' },
]

const linkClass = ({ isActive }: { isActive: boolean }) =>
  `rounded-lg px-3 py-2 text-sm font-medium transition-colors ${
    isActive ? 'bg-accent/15 text-accent' : 'text-muted hover:text-foreground'
  }`

function Navigation({ className }: { className: string }) {
  return (
    <nav className={className} aria-label="Основная навигация">
      {NAV.map(({ to, label, page, end }) => (
        <NavLink
          key={to}
          to={to}
          end={end}
          className={linkClass}
          onMouseEnter={() => preloadPage(page)}
          onFocus={() => preloadPage(page)}
        >
          {label}
        </NavLink>
      ))}
    </nav>
  )
}

export function Layout() {
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
            <ThemeToggle />
            <span
              className="grid size-8 place-items-center rounded-full bg-surface text-sm font-bold ring-1 ring-border"
              aria-hidden="true"
            >
              {session.login.slice(0, 1).toUpperCase()}
            </span>
            <span className="hidden text-sm text-muted md:inline">{session.login}</span>
            <Button size="sm" variant="ghost" onPress={signOut} aria-label="Выйти из аккаунта">
              <ArrowRightFromSquare />
              <span className="hidden sm:inline">Выйти</span>
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
