import ChevronRight from '@gravity-ui/icons/ChevronRight'
import { Chip } from '@heroui/react/chip'
import { Spinner } from '@heroui/react/spinner'
import { Link } from 'react-router-dom'

import { usePartsIndex } from '@/features/catalog/hooks'
import { formatPrice } from '@/shared/lib/format'
import { ErrorAlert } from '@/shared/ui/ErrorAlert'
import { errorMessage } from '@/shared/api/http'
import type { Order } from '@/shared/api/types'

import { groupPartUuids } from './OrderComposition'
import { useOrders } from './hooks'
import { STATUS_LABEL, STATUS_TONE } from './status'

export function OrdersPage() {
  const { data: orders, error, isLoading } = useOrders()
  const partsIndex = usePartsIndex()

  const hasNoOrders = !isLoading && (orders === undefined || orders.length === 0)

  return (
    <div className="mx-auto flex max-w-3xl flex-col gap-6">
      <h1 className="text-3xl font-bold tracking-tight">Мои заказы</h1>

      <ErrorAlert message={error ? errorMessage(error, 'Не удалось загрузить заказы') : null} />

      {isLoading ? (
        <div className="flex justify-center py-12">
          <Spinner aria-label="Загрузка заказов" />
        </div>
      ) : null}

      {hasNoOrders && !error ? (
        <div className="rounded-2xl border border-dashed border-border p-10 text-center">
          <p className="font-medium">Заказов пока нет</p>
          <p className="mt-1 text-sm text-muted">
            <Link className="text-accent underline underline-offset-4" to="/">
              Выберите детали в каталоге
            </Link>{' '}
            и оформите первый заказ.
          </p>
        </div>
      ) : null}

      <ul className="flex flex-col gap-3">
        {orders?.map((order, index) => (
          <li key={order.order_uuid} className="rise" style={{ animationDelay: `${Math.min(index, 8) * 40}ms` }}>
            <OrderRow order={order} partNames={describeParts(order, partsIndex)} />
          </li>
        ))}
      </ul>
    </div>
  )
}

function describeParts(order: Order, index: ReturnType<typeof usePartsIndex>): string | null {
  if (!index || index.size === 0) return null
  return groupPartUuids(order.part_uuids)
    .map(({ id, quantity }) => {
      const part = index.get(id)
      return part ? `${part.name}${quantity > 1 ? ` × ${quantity}` : ''}` : 'Деталь недоступна'
    })
    .join(', ')
}

function OrderRow({ order, partNames }: { order: Order; partNames: string | null }) {
  const count = order.part_uuids.length
  return (
    <Link
      to={`/orders/${order.order_uuid}`}
      className="lift flex items-center justify-between gap-4 rounded-2xl border border-border bg-surface p-4 hover:border-accent/60"
    >
      <div className="min-w-0">
        <p className="truncate font-mono text-sm">{order.order_uuid}</p>
        {partNames ? <p className="mt-1 truncate text-sm">{partNames}</p> : null}
        <p className="mt-0.5 text-sm text-muted">
          Деталей: {count}
          {order.total_price === undefined ? null : (
            <>
              {' · '}
              <span className="font-medium tabular-nums text-foreground">{formatPrice(order.total_price)}</span>
            </>
          )}
        </p>
      </div>
      <span className="flex shrink-0 items-center gap-2">
        <Chip size="sm" color={STATUS_TONE[order.status]}>
          {STATUS_LABEL[order.status]}
        </Chip>
        <ChevronRight className="size-4 text-muted" aria-hidden="true" />
      </span>
    </Link>
  )
}
