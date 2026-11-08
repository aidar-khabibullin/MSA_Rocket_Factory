import type { Part } from '../model/part'
import type { ListPartsResponseDto, PartDto } from './part.dto'

/** int64 приходит строкой; нечисловое значение считаем нулевым остатком, чтобы деталь нельзя было заказать. */
const toStock = (value: string): number => {
  const parsed = Number(value)
  return Number.isFinite(parsed) ? parsed : 0
}

export const partFromDto = (dto: PartDto): Part => ({
  uuid: dto.part_uuid,
  name: dto.name,
  description: dto.description,
  price: dto.price,
  stockQuantity: toStock(dto.stock_quantity),
  category: dto.category,
  manufacturer: dto.manufacturer?.name ?? '',
  tags: dto.tags ?? [],
})

/** Пустой каталог proto-JSON может прислать без поля parts. */
export const partsFromDto = (dto: ListPartsResponseDto): Part[] => (dto.parts ?? []).map(partFromDto)
