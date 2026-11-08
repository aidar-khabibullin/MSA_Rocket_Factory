/** Деталь в терминах фронтенда: не зависит от формата ответа бэкенда (см. repository/part.adapter.ts). */
export interface Part {
  uuid: string
  name: string
  description: string
  price: number
  stockQuantity: number
  category: string
  manufacturer: string
  tags: string[]
}
