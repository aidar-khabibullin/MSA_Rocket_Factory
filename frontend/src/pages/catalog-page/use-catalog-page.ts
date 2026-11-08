import { useDeferredValue, useEffect, useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'

import { useDebouncedValue } from '@/common/lib/use-debounced-value'
import { useUrlParam } from '@/common/lib/use-url-param'
import { useRequiredSession } from '@/modules/auth'
import {
  categoryFromParam,
  categoryToParam,
  countByCategory,
  filterParts,
  sumSelected,
  useCartStore,
  useParts,
  type CategoryKey,
  type Part,
} from '@/modules/catalog'
import { rememberOrder, useCreateOrder } from '@/modules/orders'

const QUERY_PARAM = 'q'
const CATEGORY_PARAM = 'category'
const URL_SYNC_DELAY_MS = 300

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
  const { mutateAsync: createOrder, isPending: isCreating, error: createError } = useCreateOrder()

  // Фильтры живут в URL (?q=…&category=…): страницу можно обновить или отправить ссылкой.
  const [urlQuery, setUrlQuery] = useUrlParam(QUERY_PARAM)
  const [urlCategory, setUrlCategory] = useUrlParam(CATEGORY_PARAM)
  const category = categoryFromParam(urlCategory)
  const setCategory = (next: CategoryKey | null) => setUrlCategory(next === null ? null : categoryToParam(next))

  // Поле ввода держит собственное состояние, чтобы набор текста не ждал обновления адреса;
  // в URL значение попадает с задержкой.
  const [query, setQuery] = useState(urlQuery ?? '')
  const debouncedQuery = useDebouncedValue(query, URL_SYNC_DELAY_MS)

  useEffect(() => {
    const next = debouncedQuery.trim() === '' ? null : debouncedQuery
    if (next !== urlQuery) setUrlQuery(next)
    // Только на изменение введённого текста: внешние изменения URL обрабатывает эффект ниже.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [debouncedQuery])

  useEffect(() => {
    // Адрес изменили снаружи (ссылка «Каталог» в шапке, ручная правка): поле подхватывает значение.
    if ((urlQuery ?? '') !== debouncedQuery.trim()) setQuery(urlQuery ?? '')
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [urlQuery])

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
    isCreating,
    isLoading: parts === undefined && !loadError,
    isFiltered: query.trim() !== '' || category !== null,
  }
}
