/**
 * Деталь в ответе GET /api/parts через Envoy (proto-JSON, snake_case). Дальше репозитория она не уходит.
 * int64 приходит строкой, производитель вложенным сообщением, категория именем enum.
 */
export interface PartDto {
  part_uuid: string
  name: string
  description: string
  price: number
  stock_quantity: string
  category: string
  manufacturer?: { name: string; country?: string; url?: string }
  tags?: string[]
}

/** Ответ GET /api/parts: список обёрнут в объект. */
export interface ListPartsResponseDto {
  parts?: PartDto[]
}
