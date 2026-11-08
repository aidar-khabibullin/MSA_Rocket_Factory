import { Link } from 'react-router-dom'

import { Logo } from '@/shared/ui/Logo'

import { ServiceStatusPanel } from './ServiceStatusPanel'

// Подробный статус сервисов нужен разработчику, поэтому показывается только в dev-сборке.
const SHOW_SERVICE_STATUS = import.meta.env.DEV
const YEAR = new Date().getFullYear()

export function Footer() {
  return (
    <footer className="border-t border-border/60 bg-background/60">
      <div
        className={`mx-auto grid max-w-6xl gap-8 px-4 py-8 sm:px-6 ${SHOW_SERVICE_STATUS ? 'md:grid-cols-3' : 'md:grid-cols-2'}`}
      >
        <div className="flex flex-col gap-3">
          <Logo />
          <p className="max-w-xs text-sm text-muted">Детали для ракет: выбор, заказ, оплата и сборка в одном месте.</p>
        </div>

        <nav className="flex flex-col gap-2 text-sm" aria-label="Разделы">
          <h2 className="font-bold">Разделы</h2>
          <Link to="/" className="w-fit text-muted hover:text-foreground">
            Каталог деталей
          </Link>
          <Link to="/orders" className="w-fit text-muted hover:text-foreground">
            Мои заказы
          </Link>
        </nav>

        {SHOW_SERVICE_STATUS ? <ServiceStatusPanel /> : null}
      </div>
      <div className="border-t border-border/60 py-4 text-center text-xs text-muted">© {YEAR} Rocket Factory</div>
    </footer>
  )
}
