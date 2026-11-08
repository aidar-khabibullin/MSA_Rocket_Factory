import { useMemo } from 'react'
import { useTranslation } from 'react-i18next'

import { usePartsIndex } from '@/modules/catalog'
import { groupPartUuids, useOrders, type Order } from '@/modules/orders'

export interface OrderRowModel {
  order: Order
  /** «Название × N, …»; null, пока каталог ещё не загружен. */
  partNames: string | null
}

/** Данные страницы заказов: список, подписи составов и признаки состояния. */
export function useOrdersPage() {
  const { t } = useTranslation('orders')
  const { data: orders, error, isLoading } = useOrders()
  const partsIndex = usePartsIndex()
  const unavailable = t('composition.unavailable')

  const rows = useMemo<OrderRowModel[]>(
    () =>
      (orders ?? []).map((order) => ({
        order,
        partNames:
          !partsIndex || partsIndex.size === 0
            ? null
            : groupPartUuids(order.partUuids)
                .map(({ id, quantity }) => {
                  const part = partsIndex.get(id)
                  return part ? `${part.name}${quantity > 1 ? ` × ${quantity}` : ''}` : unavailable
                })
                .join(', '),
      })),
    [orders, partsIndex, unavailable],
  )

  return { rows, error, isLoading, hasNoOrders: !isLoading && rows.length === 0 }
}
