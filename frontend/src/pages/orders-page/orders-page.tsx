import { Spinner } from '@heroui/react/spinner'
import { useTranslation } from 'react-i18next'
import { Link } from 'react-router-dom'

import { errorMessage } from '@/common/api/http'
import { ErrorAlert } from '@/common/components/error-alert'

import { OrderRow } from './order-row'
import { useOrdersPage } from './use-orders-page'

export const OrdersPage = () => {
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
