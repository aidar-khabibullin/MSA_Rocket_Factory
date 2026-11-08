import { ORDER_STATUSES, PAYMENT_METHODS, type Order, type OrderStatus, type PaymentMethod } from '../model/order'
import { isRealUuid } from '../model/status'
import type { CreatedOrderDto, OrderDto } from './order.dto'

const isPaymentMethod = (value: string | undefined): value is PaymentMethod =>
  PAYMENT_METHODS.some((method) => method === value)

const isOrderStatus = (value: string): value is OrderStatus => ORDER_STATUSES.some((status) => status === value)

export const orderFromDto = (dto: OrderDto): Order => ({
  uuid: dto.order_uuid,
  userUuid: dto.user_uuid,
  partUuids: dto.part_uuids,
  totalPrice: dto.total_price,
  // Бэкенд отдаёт нулевой UUID и 'UNKNOWN', пока заказ не оплачен: в модели это просто «нет значения».
  transactionUuid: isRealUuid(dto.transaction_uuid) ? dto.transaction_uuid : undefined,
  paymentMethod: isPaymentMethod(dto.payment_method) ? dto.payment_method : undefined,
  // Неизвестный статус (бэкенд новее фронта) трактуем как ожидание оплаты: он не опрашивается и не даёт действий.
  status: isOrderStatus(dto.status) ? dto.status : 'PENDING_PAYMENT',
})

export const createdOrderFromDto = (dto: CreatedOrderDto): { uuid: string; totalPrice: number | undefined } => ({
  uuid: dto.order_uuid,
  totalPrice: dto.total_price,
})
