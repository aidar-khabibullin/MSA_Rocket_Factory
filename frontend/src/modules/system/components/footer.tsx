import { useTranslation } from 'react-i18next'
import { Link } from 'react-router-dom'

import { Logo } from '@/common/components/logo'

import { ServiceStatusPanel } from './service-status-panel'

// Подробный статус сервисов нужен разработчику, поэтому показывается только в dev-сборке.
const SHOW_SERVICE_STATUS = import.meta.env.DEV
const YEAR = new Date().getFullYear()

export function Footer() {
  const { t } = useTranslation('system')
  return (
    <footer className="border-t border-border/60 bg-background/60">
      <div
        className={`mx-auto grid max-w-6xl gap-8 px-4 py-8 sm:px-6 ${SHOW_SERVICE_STATUS ? 'md:grid-cols-3' : 'md:grid-cols-2'}`}
      >
        <div className="flex flex-col gap-3">
          <Logo />
          <p className="max-w-xs text-sm text-muted">{t('footer.tagline')}</p>
        </div>

        <nav className="flex flex-col gap-2 text-sm" aria-label={t('footer.sections')}>
          <h2 className="font-bold">{t('footer.sections')}</h2>
          <Link to="/" className="w-fit text-muted hover:text-foreground">
            {t('footer.catalog')}
          </Link>
          <Link to="/orders" className="w-fit text-muted hover:text-foreground">
            {t('footer.orders')}
          </Link>
        </nav>

        {SHOW_SERVICE_STATUS ? <ServiceStatusPanel /> : null}
      </div>
      <div className="border-t border-border/60 py-4 text-center text-xs text-muted">© {YEAR} Rocket Factory</div>
    </footer>
  )
}
