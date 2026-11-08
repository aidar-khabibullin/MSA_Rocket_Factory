import ChevronRight from '@gravity-ui/icons/ChevronRight'
import { Chip } from '@heroui/react/chip'
import { useTranslation } from 'react-i18next'
import { Link } from 'react-router-dom'

import { formatPrice } from '@/common/lib/format'
import { STATUS_TONE, type Order } from '@/modules/orders'

interface OrderRowProps {
  order: Order
  partNames: string | null
}

export const OrderRow = ({ order, partNames }: OrderRowProps) => {
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
