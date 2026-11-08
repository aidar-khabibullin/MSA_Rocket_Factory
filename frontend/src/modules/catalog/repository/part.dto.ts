/** Ответ GET /api/parts: форма, как её отдаёт бэкенд (snake_case). Дальше репозитория она не уходит. */
export interface PartDto {
  part_uuid: string
  name: string
  description: string
  price: number
  stock_quantity: number
  category: string
  manufacturer: string
  tags?: string[]
}
