import useSWR from 'swr'
import useSWRMutation from 'swr/mutation'

import { useRequiredSession } from '@/features/auth/session'
import { ordersApi } from '@/shared/api/endpoints'
import { swrKeys } from '@/shared/api/keys'
import type { Order, PaymentMethod } from '@/shared/api/types'

import { useOrderHistory } from './history'
import { isAwaitingAssembly } from './status'

const POLL_INTERVAL_MS = 3000

/** Заказ с автоопросом: пока идёт сборка, статус обновляется сам, после ASSEMBLED опрос прекращается. */
export function useOrder(orderUuid: string) {
  const { sessionUuid } = useRequiredSession()
  return useSWR(swrKeys.order(sessionUuid, orderUuid), ([, session, id]) => ordersApi.get(session, id), {
    refreshInterval: (order) => (isAwaitingAssembly(order?.status) ? POLL_INTERVAL_MS : 0),
  })
}

/** Все заказы пользователя: запросы независимы, поэтому идут параллельно (правило async-parallel). */
export function useOrders() {
  const { sessionUuid, userUuid } = useRequiredSession()
  const orderUuids = useOrderHistory(userUuid)

  return useSWR(
    orderUuids.length > 0 ? swrKeys.orders(sessionUuid, orderUuids) : null,
    async ([, session, ...ids]): Promise<Order[]> => {
      const results = await Promise.allSettled(ids.map((id) => ordersApi.get(session, id)))
      return results.flatMap((result) => (result.status === 'fulfilled' ? [result.value] : []))
    },
    { revalidateOnFocus: false },
  )
}

export function useCreateOrder() {
  const { sessionUuid, userUuid } = useRequiredSession()
  return useSWRMutation('create-order', (_key: string, { arg }: { arg: readonly string[] }) =>
    ordersApi.create(sessionUuid, userUuid, [...arg]),
  )
}

/** После оплаты и отмены тот же ключ перечитывается с сервера: источник правды остаётся бэкенд. */
export function usePayOrder(orderUuid: string) {
  const { sessionUuid } = useRequiredSession()
  return useSWRMutation(
    swrKeys.order(sessionUuid, orderUuid),
    (_key, { arg }: { arg: PaymentMethod }) => ordersApi.pay(sessionUuid, orderUuid, arg),
    { populateCache: false },
  )
}

export function useCancelOrder(orderUuid: string) {
  const { sessionUuid } = useRequiredSession()
  return useSWRMutation(swrKeys.order(sessionUuid, orderUuid), () => ordersApi.cancel(sessionUuid, orderUuid), {
    populateCache: false,
  })
}
