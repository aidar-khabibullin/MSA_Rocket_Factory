import ArrowLeft from '@gravity-ui/icons/ArrowLeft'
import { Button } from '@heroui/react/button'
import { Spinner } from '@heroui/react/spinner'
import { useTranslation } from 'react-i18next'
import { Link, useParams } from 'react-router-dom'

import { errorMessage } from '@/common/api/http'
import type { Order } from '@/modules/orders'
import { formatPrice } from '@/common/lib/format'
import { AsyncGate } from '@/common/components/async-gate'
import { ErrorAlert } from '@/common/components/error-alert'
import { ConfirmDialog } from '@/common/components/confirm-dialog'

import { isAwaitingAssembly, OrderComposition, OrderStepper, PaymentMethodPicker, useOrder } from '@/modules/orders'

import { usePaymentPanel } from './use-payment-panel'

export function OrderPage() {
  const { t } = useTranslation('orders')
  const { orderId = '' } = useParams()
  const { data: order, error, isLoading } = useOrder(orderId)

  return (
    <AsyncGate
      data={order}
      isLoading={isLoading}
      loader={
        <div className="flex justify-center py-16">
          <Spinner aria-label={t('details.loading')} />
        </div>
      }
      fallback={
        <div className="mx-auto flex max-w-2xl flex-col gap-4">
          <BackLink />
          <ErrorAlert message={errorMessage(error, t('details.notFound'))} />
        </div>
      }
    >
      {(loaded) => <OrderDetails order={loaded} />}
    </AsyncGate>
  )
}

function OrderDetails({ order }: { order: Order }) {
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
          <Row label={t('details.partsInOrder')} value={String(order.partUuids.length)} />
          <Row
            label={t('details.total')}
            value={order.totalPrice === undefined ? '—' : formatPrice(order.totalPrice)}
            strong
          />
          {paymentMethod ? <Row label={t('details.paymentMethod')} value={t(`payment.${paymentMethod}`)} /> : null}
          {order.transactionUuid ? <Row label={t('details.transaction')} value={order.transactionUuid} mono /> : null}
        </dl>
      </section>

      {order.status === 'PENDING_PAYMENT' ? <PaymentPanel order={order} /> : null}
    </div>
  )
}

function PaymentPanel({ order }: { order: Order }) {
  const { t } = useTranslation('orders')
  const panel = usePaymentPanel(order.uuid)

  return (
    <section className="flex flex-col gap-4 rounded-2xl border border-accent/40 bg-surface p-6" aria-label={t('pay.region')}>
      <PaymentMethodPicker value={panel.method} onChange={panel.setMethod} />

      <ErrorAlert message={panel.error ? errorMessage(panel.error, t('pay.failed')) : null} />

      <div className="flex flex-wrap gap-2">
        <Button className="flex-1" isDisabled={panel.isBusy} onPress={() => void panel.payOrder()}>
          {panel.isPaying ? <Spinner size="sm" color="current" /> : null}
          {t('pay.pay')}
          {order.totalPrice === undefined ? '' : ` ${formatPrice(order.totalPrice)}`}
        </Button>
        <Button variant="danger-soft" isDisabled={panel.isBusy} onPress={() => panel.setCancelDialogOpen(true)}>
          {t('pay.cancel')}
        </Button>
      </div>

      <ConfirmDialog
        isOpen={panel.isCancelDialogOpen}
        onOpenChange={panel.setCancelDialogOpen}
        title={t('pay.confirmTitle')}
        description={t('pay.confirmText')}
        confirmLabel={t('pay.confirmAction')}
        cancelLabel={t('pay.keep')}
        onConfirm={() => void panel.confirmCancel()}
        isPending={panel.isCancelling}
      />
    </section>
  )
}

function BackLink() {
  const { t } = useTranslation('orders')
  return (
    <Link to="/orders" className="inline-flex w-fit items-center gap-1.5 text-sm text-muted hover:text-foreground">
      <ArrowLeft className="size-4" aria-hidden="true" /> {t('details.back')}
    </Link>
  )
}

function Row({ label, value, strong, mono }: { label: string; value: string; strong?: boolean; mono?: boolean }) {
  return (
    <div className="flex flex-col gap-0.5">
      <dt className="text-muted">{label}</dt>
      <dd className={`break-all ${strong ? 'text-lg font-bold tabular-nums' : ''} ${mono ? 'font-mono text-xs' : ''}`}>
        {value}
      </dd>
    </div>
  )
}
