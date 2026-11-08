import ArrowLeft from '@gravity-ui/icons/ArrowLeft'
import { Button } from '@heroui/react/button'
import { Spinner } from '@heroui/react/spinner'
import { useState } from 'react'
import { Link, useParams } from 'react-router-dom'

import { errorMessage } from '@/shared/api/http'
import type { Order, PaymentMethod } from '@/shared/api/types'
import { formatPrice } from '@/shared/lib/format'
import { ErrorAlert } from '@/shared/ui/ErrorAlert'

import { useCancelOrder, useOrder, usePayOrder } from './hooks'
import { OrderComposition } from './OrderComposition'
import { OrderStepper } from './OrderStepper'
import { PaymentMethodPicker } from './PaymentMethodPicker'
import { isAwaitingAssembly, isRealUuid, PAYMENT_LABEL, STATUS_LABEL } from './status'

export function OrderPage() {
  const { orderId = '' } = useParams()
  const { data: order, error, isLoading } = useOrder(orderId)

  if (isLoading) {
    return (
      <div className="flex justify-center py-16">
        <Spinner aria-label="Загрузка заказа" />
      </div>
    )
  }

  if (!order) {
    return (
      <div className="mx-auto flex max-w-2xl flex-col gap-4">
        <BackLink />
        <ErrorAlert message={errorMessage(error, 'Заказ не найден')} />
      </div>
    )
  }

  return <OrderDetails order={order} />
}

function OrderDetails({ order }: { order: Order }) {
  const awaitingAssembly = isAwaitingAssembly(order.status)
  const paymentMethod = order.payment_method && order.payment_method !== 'UNKNOWN' ? order.payment_method : null

  return (
    <div className="rise mx-auto flex max-w-2xl flex-col gap-6">
      <BackLink />

      <header className="flex flex-col gap-1">
        <p className="text-sm text-muted">Заказ</p>
        <h1 className="break-all font-mono text-lg font-bold sm:text-xl">{order.order_uuid}</h1>
        <p className="text-sm text-muted">{STATUS_LABEL[order.status]}</p>
      </header>

      <section className="rounded-2xl border border-border bg-surface p-6" aria-label="Статус заказа">
        <OrderStepper status={order.status} />
        {awaitingAssembly ? (
          <output className="mt-5 flex items-center justify-center gap-2 text-center text-sm text-muted">
            <Spinner size="sm" /> Ракета собирается, обычно это занимает 5–15 секунд.
          </output>
        ) : null}
        {order.status === 'ASSEMBLED' ? (
          <output className="mt-5 block text-center text-sm font-medium text-success">
            Ваша ракета готова. Уведомление отправлено в Telegram, если он подключён.
          </output>
        ) : null}
      </section>

      <OrderComposition partUuids={order.part_uuids} />

      <section className="rounded-2xl border border-border bg-surface p-6" aria-label="Сведения о заказе">
        <dl className="grid gap-4 text-sm sm:grid-cols-2">
          <Row label="Деталей в заказе" value={String(order.part_uuids.length)} />
          <Row
            label="Сумма"
            value={order.total_price === undefined ? '—' : formatPrice(order.total_price)}
            strong
          />
          {paymentMethod ? <Row label="Способ оплаты" value={PAYMENT_LABEL[paymentMethod]} /> : null}
          {isRealUuid(order.transaction_uuid) ? <Row label="Транзакция" value={order.transaction_uuid} mono /> : null}
        </dl>
      </section>

      {order.status === 'PENDING_PAYMENT' ? <PaymentPanel order={order} /> : null}
    </div>
  )
}

function PaymentPanel({ order }: { order: Order }) {
  const [method, setMethod] = useState<PaymentMethod>('CARD')
  const pay = usePayOrder(order.order_uuid)
  const cancel = useCancelOrder(order.order_uuid)

  const isPending = pay.isMutating || cancel.isMutating
  const error = pay.error ?? cancel.error

  async function run(action: () => Promise<unknown>) {
    try {
      await action()
    } catch {
      /* ошибка сохранена в состоянии мутации и показывается ниже */
    }
  }

  return (
    <section className="flex flex-col gap-4 rounded-2xl border border-accent/40 bg-surface p-6" aria-label="Оплата">
      <PaymentMethodPicker value={method} onChange={setMethod} />

      <ErrorAlert message={error ? errorMessage(error, 'Операция не выполнена') : null} />

      <div className="flex flex-wrap gap-2">
        <Button className="flex-1" isDisabled={isPending} onPress={() => run(() => pay.trigger(method))}>
          {pay.isMutating ? <Spinner size="sm" color="current" /> : null}
          Оплатить{order.total_price === undefined ? '' : ` ${formatPrice(order.total_price)}`}
        </Button>
        <Button variant="danger-soft" isDisabled={isPending} onPress={() => run(() => cancel.trigger())}>
          Отменить заказ
        </Button>
      </div>
    </section>
  )
}

function BackLink() {
  return (
    <Link to="/orders" className="inline-flex w-fit items-center gap-1.5 text-sm text-muted hover:text-foreground">
      <ArrowLeft className="size-4" aria-hidden="true" /> Все заказы
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
