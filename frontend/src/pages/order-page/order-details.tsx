import { Spinner } from '@heroui/react/spinner'
import { useTranslation } from 'react-i18next'

import { formatPrice } from '@/common/lib/format'
import { isAwaitingAssembly, OrderComposition, OrderStepper, type Order } from '@/modules/orders'

import { BackLink } from './back-link'
import { InfoRow } from './info-row'
import { PaymentPanel } from './payment-panel'

interface OrderDetailsProps {
  order: Order
}

export const OrderDetails = ({ order }: OrderDetailsProps) => {
  const { t } = useTranslation('orders')
  const awaitingAssembly = isAwaitingAssembly(order.status)
  const { paymentMethod } = order

  return (
    <div className="rise mx-auto flex max-w-2xl flex-col gap-6">
      <BackLink />

      <header className="flex flex-col gap-1">
        <p className="text-sm text-muted">{t('details.order')}</p>
        <h1 className="break-all font-mono text-lg font-bold sm:text-xl">{order.uuid}</h1>
        <p className="text-sm text-muted">{t(`status.${order.status}`)}</p>
      </header>

      <section className="rounded-2xl border border-border bg-surface p-6" aria-label={t('details.statusRegion')}>
        <OrderStepper status={order.status} />
        {awaitingAssembly ? (
          <output className="mt-5 flex items-center justify-center gap-2 text-center text-sm text-muted">
            <Spinner size="sm" /> {t('details.assembling')}
          </output>
        ) : null}
        {order.status === 'ASSEMBLED' ? (
          <output className="mt-5 block text-center text-sm font-medium text-success">
            {t('details.ready')}
          </output>
        ) : null}
      </section>

      <OrderComposition partUuids={order.partUuids} />

      <section className="rounded-2xl border border-border bg-surface p-6" aria-label={t('details.infoRegion')}>
        <dl className="grid gap-4 text-sm sm:grid-cols-2">
          <InfoRow label={t('details.partsInOrder')} value={String(order.partUuids.length)} />
          <InfoRow
            label={t('details.total')}
            value={order.totalPrice === undefined ? '—' : formatPrice(order.totalPrice)}
            strong
          />
          {paymentMethod ? <InfoRow label={t('details.paymentMethod')} value={t(`payment.${paymentMethod}`)} /> : null}
          {order.transactionUuid ? (
            <InfoRow label={t('details.transaction')} value={order.transactionUuid} mono />
          ) : null}
        </dl>
      </section>

      {order.status === 'PENDING_PAYMENT' ? <PaymentPanel order={order} /> : null}
    </div>
  )
}
