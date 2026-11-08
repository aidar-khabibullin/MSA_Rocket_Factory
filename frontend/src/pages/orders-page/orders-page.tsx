import ChevronRight from '@gravity-ui/icons/ChevronRight'
import { Chip } from '@heroui/react/chip'
import { Spinner } from '@heroui/react/spinner'
import { useTranslation } from 'react-i18next'
import { Link } from 'react-router-dom'

import { errorMessage } from '@/common/api/http'
import { ErrorAlert } from '@/common/components/error-alert'
import { formatPrice } from '@/common/lib/format'
import { STATUS_TONE, type Order } from '@/modules/orders'

import { useOrdersPage } from './use-orders-page'

export function OrdersPage() {
  const { t } = useTranslation('orders')
  const { rows, error, isLoading, hasNoOrders } = useOrdersPage()

  return (
    <div className="mx-auto flex max-w-3xl flex-col gap-6">
      <h1 className="text-3xl font-bold tracking-tight">{t('list.title')}</h1>

      <ErrorAlert message={error ? errorMessage(error, t('list.error')) : null} />

      {isLoading ? (
        <div className="flex justify-center py-12">
          <Spinner aria-label={t('list.loading')} />
        </div>
      ) : null}

      {hasNoOrders && !error ? (
        <div className="rounded-2xl border border-dashed border-border p-10 text-center">
          <p className="font-medium">{t('list.emptyTitle')}</p>
          <p className="mt-1 text-sm text-muted">
            <Link className="text-accent underline underline-offset-4" to="/">
              {t('list.emptyLink')}
            </Link>{' '}
            {t('list.emptyTail')}
          </p>
        </div>
      ) : null}

      <ul className="flex flex-col gap-3">
        {rows.map(({ order, partNames }, index) => (
          <li key={order.uuid} className="rise" style={{ animationDelay: `${Math.min(index, 8) * 40}ms` }}>
            <OrderRow order={order} partNames={partNames} />
          </li>
        ))}
      </ul>
    </div>
  )
}

function OrderRow({ order, partNames }: { order: Order; partNames: string | null }) {
  const { t } = useTranslation('orders')
  const count = order.partUuids.length
  return (
    <Link
      to={`/orders/${order.uuid}`}
      className="lift flex items-center justify-between gap-4 rounded-2xl border border-border bg-surface p-4 hover:border-accent/60"
    >
      <div className="min-w-0">
        <p className="truncate font-mono text-sm">{order.uuid}</p>
        {partNames ? <p className="mt-1 truncate text-sm">{partNames}</p> : null}
        <p className="mt-0.5 text-sm text-muted">
          {t('list.partsCount', { n: count })}
          {order.totalPrice === undefined ? null : (
            <>
              {' · '}
              <span className="font-medium tabular-nums text-foreground">{formatPrice(order.totalPrice)}</span>
            </>
          )}
        </p>
      </div>
      <span className="flex shrink-0 items-center gap-2">
        <Chip size="sm" color={STATUS_TONE[order.status]}>
          {t(`status.${order.status}`)}
        </Chip>
        <ChevronRight className="size-4 text-muted" aria-hidden="true" />
      </span>
    </Link>
  )
}
