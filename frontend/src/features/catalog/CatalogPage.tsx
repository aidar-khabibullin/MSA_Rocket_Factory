import Magnifier from '@gravity-ui/icons/Magnifier'
import { Button } from '@heroui/react/button'
import { Input } from '@heroui/react/input'
import { useDeferredValue, useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'

import { useRequiredSession } from '@/features/auth/session'
import { useCreateOrder } from '@/features/orders/hooks'
import { rememberOrder } from '@/features/orders/history'
import { errorMessage } from '@/shared/api/http'
import type { Part } from '@/shared/api/types'
import { ErrorAlert } from '@/shared/ui/ErrorAlert'

import { CartBar } from './CartBar'
import type { CategoryKey } from './categories'
import { CategoryFilter } from './CategoryFilter'
import { countByCategory, filterParts, sumSelected } from './filters'
import { useCartStore, useParts } from './hooks'
import { PartCard, PartCardSkeleton } from './PartCard'

const SKELETONS = [0, 1, 2] as const
// Стабильная ссылка: `parts ?? []` создавал бы новый массив на каждый рендер и ломал бы useMemo.
const NO_PARTS: readonly Part[] = []

export function CatalogPage() {
  const { userUuid } = useRequiredSession()
  const navigate = useNavigate()
  const { data: parts, error: loadError } = useParts()
  const selected = useCartStore((state) => state.selected)
  const toggle = useCartStore((state) => state.toggle)
  const clear = useCartStore((state) => state.clear)
  const { trigger: createOrder, isMutating, error: createError } = useCreateOrder()

  const [query, setQuery] = useState('')
  const [category, setCategory] = useState<CategoryKey | null>(null)
  // Ввод остаётся отзывчивым: тяжёлая фильтрация идёт по отложенному значению (правило rerender-use-deferred-value).
  const deferredQuery = useDeferredValue(query)

  const all = parts ?? NO_PARTS
  const visible = useMemo(
    () => filterParts(all, { query: deferredQuery, category }),
    [all, deferredQuery, category],
  )
  const counts = useMemo(() => countByCategory(all), [all])
  const total = useMemo(() => sumSelected(all, selected), [all, selected])

  async function checkout() {
    try {
      const { order_uuid: orderUuid } = await createOrder([...selected])
      rememberOrder(userUuid, orderUuid)
      clear()
      void navigate(`/orders/${orderUuid}`)
    } catch {
      /* ошибка уже сохранена в createError и показывается ниже */
    }
  }

  const isLoading = parts === undefined && !loadError
  const isFiltered = query.trim() !== '' || category !== null

  return (
    <div className="flex flex-col gap-8 pb-28">
      <section className="flex flex-col gap-2">
        <h1 className="text-3xl font-bold tracking-tight sm:text-4xl">Соберите свою ракету</h1>
        <p className="max-w-2xl text-muted">
          Выберите детали из каталога, оформите заказ и оплатите. Сборка начнётся сразу после оплаты.
        </p>
      </section>

      <section className="flex flex-col gap-4" aria-label="Поиск и фильтры">
        <div className="relative max-w-md">
          <Magnifier
            className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted"
            aria-hidden="true"
          />
          <Input
            aria-label="Поиск деталей"
            placeholder="Поиск по названию или описанию"
            value={query}
            onChange={(event) => setQuery(event.target.value)}
            className="w-full pl-9"
          />
        </div>
        <CategoryFilter value={category} onChange={setCategory} counts={counts} total={parts?.length} />
      </section>

      <ErrorAlert message={loadError ? errorMessage(loadError, 'Не удалось загрузить каталог') : null} />
      <ErrorAlert message={createError ? errorMessage(createError, 'Не удалось создать заказ') : null} />

      {isLoading ? (
        <ul className="grid gap-5 sm:grid-cols-2 lg:grid-cols-3" aria-label="Загрузка каталога">
          {SKELETONS.map((key) => (
            <li key={key}>
              <PartCardSkeleton />
            </li>
          ))}
        </ul>
      ) : null}

      {parts !== undefined && visible.length === 0 ? (
        <div className="rounded-2xl border border-dashed border-border p-10 text-center">
          <p className="font-medium">{isFiltered ? 'Ничего не найдено' : 'Каталог пуст'}</p>
          {isFiltered ? (
            <>
              <p className="mt-1 text-sm text-muted">Измените запрос или сбросьте фильтр.</p>
              <Button
                className="mt-4"
                variant="outline"
                onPress={() => {
                  setQuery('')
                  setCategory(null)
                }}
              >
                Сбросить фильтры
              </Button>
            </>
          ) : null}
        </div>
      ) : null}

      <ul className="grid gap-5 sm:grid-cols-2 lg:grid-cols-3">
        {visible.map((part, index) => (
          <li key={part.part_uuid}>
            <PartCard part={part} selected={selected.has(part.part_uuid)} onToggle={toggle} index={index} />
          </li>
        ))}
      </ul>

      {selected.size > 0 ? (
        <CartBar count={selected.size} total={total} isPending={isMutating} onClear={clear} onCheckout={checkout} />
      ) : null}
    </div>
  )
}
