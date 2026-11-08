export const ORDER_STATUSES = ['PENDING_PAYMENT', 'PAID', 'ASSEMBLED', 'CANCELLED'] as const
export type OrderStatus = (typeof ORDER_STATUSES)[number]

export const PAYMENT_METHODS = ['CARD', 'SBP', 'CREDIT_CARD', 'INVESTOR_MONEY'] as const
export type PaymentMethod = (typeof PAYMENT_METHODS)[number]

export interface Order {
  uuid: string
  userUuid: string
  partUuids: string[]
  totalPrice?: number
  /** Есть только у оплаченного заказа. */
  transactionUuid?: string
  paymentMethod?: PaymentMethod
  status: OrderStatus
}
