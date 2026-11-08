import { useTranslation } from 'react-i18next'
import { NavLink } from 'react-router-dom'

import { preloadPage, type PageName } from '@/app/routes'

interface NavigationProps {
  className: string
}

const NAV: ReadonlyArray<{ to: string; labelKey: 'nav.catalog' | 'nav.orders'; page: PageName; end?: boolean }> = [
  { to: '/', labelKey: 'nav.catalog', page: 'catalog', end: true },
  { to: '/orders', labelKey: 'nav.orders', page: 'orders' },
]

const linkClass = ({ isActive }: { isActive: boolean }) =>
  `rounded-lg px-3 py-2 text-sm font-medium transition-colors ${
    isActive ? 'bg-accent/15 text-accent' : 'text-muted hover:text-foreground'
  }`

export const Navigation = ({ className }: NavigationProps) => {
  const { t } = useTranslation()
  return (
    <nav className={className} aria-label={t('nav.main')}>
      {NAV.map(({ to, labelKey, page, end }) => (
        <NavLink
          key={to}
          to={to}
          end={end}
          className={linkClass}
          onMouseEnter={() => preloadPage(page)}
          onFocus={() => preloadPage(page)}
        >
          {t(labelKey)}
        </NavLink>
      ))}
    </nav>
  )
}
