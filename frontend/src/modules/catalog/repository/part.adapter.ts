import type { Part } from '../model/part'
import type { PartDto } from './part.dto'

export const partFromDto = (dto: PartDto): Part => ({
  uuid: dto.part_uuid,
  name: dto.name,
  description: dto.description,
  price: dto.price,
  stockQuantity: dto.stock_quantity,
  category: dto.category,
  manufacturer: dto.manufacturer,
  tags: dto.tags ?? [],
})
