import type { Part } from './part'

import type { CategoryKey } from './categories'

export interface PartsFilter {
  query: string
  category: CategoryKey | null
}

/** Фильтрация за один проход: категория и текст проверяются вместе. */
export function filterParts(parts: readonly Part[], { query, category }: PartsFilter): Part[] {
  const needle = query.trim().toLowerCase()
  return parts.filter(
    (part) =>
      (category === null || part.category === category) &&
      (needle === '' || part.name.toLowerCase().includes(needle) || part.description.toLowerCase().includes(needle)),
  )
}

export function countByCategory(parts: readonly Part[]): ReadonlyMap<string, number> {
  const counts = new Map<string, number>()
  for (const part of parts) counts.set(part.category, (counts.get(part.category) ?? 0) + 1)
  return counts
}

/** Сумма цен выбранных деталей; Set даёт O(1) на проверку, а не O(n) на каждую деталь. */
export function sumSelected(parts: readonly Part[], selected: ReadonlySet<string>): number {
  let total = 0
  for (const part of parts) if (selected.has(part.uuid)) total += part.price
  return total
}

export type StockLevel = 'out' | 'low' | 'ok'

export const LOW_STOCK_THRESHOLD = 5

export function stockLevel(quantity: number): StockLevel {
  if (quantity <= 0) return 'out'
  return quantity <= LOW_STOCK_THRESHOLD ? 'low' : 'ok'
}
