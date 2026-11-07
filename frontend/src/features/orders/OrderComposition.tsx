import { usePartsIndex } from '@/features/catalog/hooks'
import { categoryMeta } from '@/features/catalog/categories'
import { formatPrice } from '@/shared/lib/format'

/** Схлопывает повторяющиеся UUID в строки «деталь × количество». */
export function groupPartUuids(partUuids: readonly string[]): Array<{ id: string; quantity: number }> {
  const counts = new Map<string, number>()
  for (const id of partUuids) counts.set(id, (counts.get(id) ?? 0) + 1)
  return Array.from(counts, ([id, quantity]) => ({ id, quantity }))
}

export function OrderComposition({ partUuids }: { partUuids: readonly string[] }) {
  const index = usePartsIndex()

  return (
    <section className="rounded-2xl border border-border bg-surface p-6" aria-label="Состав заказа">
      <h2 className="mb-4 font-bold">Состав заказа</h2>
      {index === undefined ? (
        <p className="text-sm text-muted">Загрузка состава…</p>
      ) : (
        <ul className="flex flex-col divide-y divide-border/60">
          {groupPartUuids(partUuids).map(({ id, quantity }) => {
            const part = index.get(id)
            const Icon = part ? categoryMeta(part.category).icon : null
            return (
              <li key={id} className="flex items-center justify-between gap-4 py-3 first:pt-0 last:pb-0">
                <div className="flex min-w-0 items-center gap-3">
                  <span
                    className="grid size-10 shrink-0 place-items-center rounded-lg bg-accent/12 text-accent"
                    aria-hidden="true"
                  >
                    {Icon ? <Icon className="size-5" /> : null}
                  </span>
                  <div className="min-w-0">
                    <p className="truncate font-medium">{part?.name ?? 'Деталь недоступна'}</p>
                    <p className="truncate text-xs text-muted">
                      {part ? `${part.manufacturer}${quantity > 1 ? ` · ${quantity} шт.` : ''}` : id}
                    </p>
                  </div>
                </div>
                {part ? (
                  <span className="shrink-0 font-medium tabular-nums">{formatPrice(part.price * quantity)}</span>
                ) : null}
              </li>
            )
          })}
        </ul>
      )}
    </section>
  )
}
