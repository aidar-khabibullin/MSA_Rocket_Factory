import { http } from '@/common/api/http'

import type { Order, PaymentMethod } from '../model/order'
import { createdOrderFromDto, orderFromDto } from './order.adapter'
import type { CreatedOrderDto, OrderDto } from './order.dto'

export const ordersApi = {
  create: async (sessionUuid: string, userUuid: string, partUuids: string[]) =>
    createdOrderFromDto(
      await http<CreatedOrderDto>('/api/v1/orders', {
        method: 'POST',
        sessionUuid,
        body: { user_uuid: userUuid, part_uuids: partUuids },
      }),
    ),

  get: async (sessionUuid: string, orderUuid: string, signal?: AbortSignal): Promise<Order> =>
    orderFromDto(await http<OrderDto>(`/api/v1/orders/${orderUuid}`, { sessionUuid, signal })),

  pay: (sessionUuid: string, orderUuid: string, paymentMethod: PaymentMethod) =>
    http<{ transaction_uuid: string }>(`/api/v1/orders/${orderUuid}/pay`, {
      method: 'POST',
      sessionUuid,
      body: { payment_method: paymentMethod },
    }),

  cancel: (sessionUuid: string, orderUuid: string) =>
    http<void>(`/api/v1/orders/${orderUuid}/cancel`, { method: 'POST', sessionUuid }),
}
