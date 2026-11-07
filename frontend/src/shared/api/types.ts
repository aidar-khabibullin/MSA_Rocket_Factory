export const ORDER_STATUSES = ['PENDING_PAYMENT', 'PAID', 'ASSEMBLED', 'CANCELLED'] as const
export type OrderStatus = (typeof ORDER_STATUSES)[number]

export const PAYMENT_METHODS = ['CARD', 'SBP', 'CREDIT_CARD', 'INVESTOR_MONEY'] as const
export type PaymentMethod = (typeof PAYMENT_METHODS)[number]

export interface Part {
  part_uuid: string
  name: string
  description: string
  price: number
  stock_quantity: number
  category: string
  manufacturer: string
  tags: string[]
}

export interface Order {
  order_uuid: string
  user_uuid: string
  part_uuids: string[]
  total_price?: number
  transaction_uuid?: string
  payment_method?: PaymentMethod | 'UNKNOWN'
  status: OrderStatus
}

export interface ServiceStatus {
  name: string
  ok: boolean
}
