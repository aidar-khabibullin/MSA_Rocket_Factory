import { useDeferredValue, useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'

import { useRequiredSession } from '@/modules/auth'
import {
  countByCategory,
  filterParts,
  sumSelected,
  useCartStore,
  useParts,
  type CategoryKey,
  type Part,
} from '@/modules/catalog'
import { rememberOrder, useCreateOrder } from '@/modules/orders'

// Стабильная ссылка: `parts ?? []` создавал бы новый массив на каждый рендер и ломал бы useMemo.
const NO_PARTS: readonly Part[] = []

/** Вся логика страницы каталога: данные, фильтры, корзина и оформление заказа. Страница остаётся разметкой. */
export function useCatalogPage() {
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
      const { uuid: orderUuid } = await createOrder([...selected])
      rememberOrder(userUuid, orderUuid)
      clear()
      void navigate(`/orders/${orderUuid}`)
    } catch {
      /* ошибка уже сохранена в createError и показывается страницей */
    }
  }

  function resetFilters() {
    setQuery('')
    setCategory(null)
  }

  return {
    parts,
    visible,
    counts,
    total,
    selected,
    toggle,
    clear,
    checkout,
    query,
    setQuery,
    category,
    setCategory,
    resetFilters,
    loadError,
    createError,
    isCreating: isMutating,
    isLoading: parts === undefined && !loadError,
    isFiltered: query.trim() !== '' || category !== null,
  }
}
