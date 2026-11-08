import ArrowRotateRight from '@gravity-ui/icons/ArrowRotateRight'
import CircleCheck from '@gravity-ui/icons/CircleCheck'
import CircleXmark from '@gravity-ui/icons/CircleXmark'
import { useTranslation } from 'react-i18next'

import { formatTime } from '@/common/lib/format'

import { useServiceStatus } from '../hooks/use-service-status'

/** Подробный статус для разработчика: показывается только в dev-сборке (см. Footer). */
export function ServiceStatusPanel() {
  const { t } = useTranslation('system')
  const { services, checkedAt, isGatewayDown, isDegraded, isChecking, refresh } = useServiceStatus()

  return (
    <section className="flex flex-col gap-3 text-sm" aria-label={t('panel.region')}>
      <div className="flex items-center justify-between gap-2">
        <h2 className="font-bold">{t('panel.title')}</h2>
        <button
          type="button"
          onClick={refresh}
          disabled={isChecking}
          aria-label={t('panel.refresh')}
          className="grid size-8 place-items-center rounded-lg text-muted hover:text-foreground disabled:opacity-50"
        >
          <ArrowRotateRight className={`size-4 ${isChecking ? 'animate-spin' : ''}`} aria-hidden="true" />
        </button>
      </div>

      {services === undefined && !isGatewayDown ? <p className="text-muted">{t('panel.checking')}</p> : null}

      {isGatewayDown ? (
        <output className="flex items-center gap-2 text-danger">
          <CircleXmark className="size-4 shrink-0" aria-hidden="true" /> {t('panel.gatewayDown')}{' '}
          <code className="font-mono">task bff:run</code>
        </output>
      ) : null}

      {services && services.length > 0 ? (
        <ul className="flex flex-col gap-1.5">
          {services.map((service) => (
            <li key={service.name} className="flex items-center justify-between gap-2">
              <span>{service.name}</span>
              {service.ok ? (
                <CircleCheck className="size-5 text-success" aria-hidden="true" />
              ) : (
                <CircleXmark className="size-5 text-danger" aria-hidden="true" />
              )}
              <span className="sr-only">{service.ok ? t('panel.up') : t('panel.down')}</span>
            </li>
          ))}
        </ul>
      ) : null}

      {checkedAt ? (
        <p className="text-xs text-muted">
          {isDegraded ? t('panel.degraded') : t('panel.healthy')} · {t('panel.checkedAt', { time: formatTime(checkedAt) })}
        </p>
      ) : null}
    </section>
  )
}
