import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'

import { useRequiredSession } from '@/modules/auth'

import { isAwaitingAssembly } from '../model/status'
import type { Order, PaymentMethod } from '../model/order'
import { ordersApi } from '../repository/orders.api'
import { ordersKeys } from '../repository/orders.keys'
import { useOrderHistory } from '../store/order-history.store'

const POLL_INTERVAL_MS = 3000

/** Заказ с автоопросом: пока идёт сборка, статус обновляется сам, после ASSEMBLED опрос прекращается. */
export function useOrder(orderUuid: string) {
  const { sessionUuid } = useRequiredSession()
  return useQuery({
    queryKey: ordersKeys.order(sessionUuid, orderUuid),
    queryFn: () => ordersApi.get(sessionUuid, orderUuid),
    refetchInterval: (query) => (isAwaitingAssembly(query.state.data?.status) ? POLL_INTERVAL_MS : false),
  })
}

/** Все заказы пользователя: запросы независимы, поэтому идут параллельно (правило async-parallel). */
export function useOrders() {
  const { sessionUuid, userUuid } = useRequiredSession()
  const orderUuids = useOrderHistory(userUuid)

  return useQuery({
    queryKey: ordersKeys.orders(sessionUuid, orderUuids),
    queryFn: async (): Promise<Order[]> => {
      const results = await Promise.allSettled(orderUuids.map((id) => ordersApi.get(sessionUuid, id)))
      return results.flatMap((result) => (result.status === 'fulfilled' ? [result.value] : []))
    },
    enabled: orderUuids.length > 0,
    refetchOnWindowFocus: false,
  })
}

export function useCreateOrder() {
  const { sessionUuid, userUuid } = useRequiredSession()
  return useMutation({
    mutationFn: (partUuids: readonly string[]) => ordersApi.create(sessionUuid, userUuid, [...partUuids]),
  })
}

/** После оплаты и отмены заказ перечитывается с сервера: источник правды остаётся бэкенд. */
export function usePayOrder(orderUuid: string) {
  const { sessionUuid } = useRequiredSession()
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (method: PaymentMethod) => ordersApi.pay(sessionUuid, orderUuid, method),
    // Возвращаем промис: мутация завершается, когда свежий заказ уже в кэше.
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ordersKeys.order(sessionUuid, orderUuid) }),
  })
}

export function useCancelOrder(orderUuid: string) {
  const { sessionUuid } = useRequiredSession()
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: () => ordersApi.cancel(sessionUuid, orderUuid),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ordersKeys.order(sessionUuid, orderUuid) }),
  })
}
