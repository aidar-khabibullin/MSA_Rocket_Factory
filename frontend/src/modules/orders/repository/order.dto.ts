/** Ответы order-сервиса в формате бэкенда (snake_case). Дальше репозитория они не уходят. */
export interface OrderDto {
  order_uuid: string
  user_uuid: string
  part_uuids: string[]
  total_price?: number
  transaction_uuid?: string
  payment_method?: string
  status: string
}

export interface CreatedOrderDto {
  order_uuid: string
  total_price?: number
}
