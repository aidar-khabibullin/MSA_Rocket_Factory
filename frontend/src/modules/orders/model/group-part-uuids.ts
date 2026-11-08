/** Схлопывает повторяющиеся UUID в строки «деталь × количество». */
export function groupPartUuids(partUuids: readonly string[]): Array<{ id: string; quantity: number }> {
  const counts = new Map<string, number>()
  for (const id of partUuids) counts.set(id, (counts.get(id) ?? 0) + 1)
  return Array.from(counts, ([id, quantity]) => ({ id, quantity }))
}
