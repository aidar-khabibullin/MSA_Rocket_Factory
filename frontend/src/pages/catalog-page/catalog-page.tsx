import Magnifier from '@gravity-ui/icons/Magnifier'
import { Button } from '@heroui/react/button'
import { Input } from '@heroui/react/input'
import { useTranslation } from 'react-i18next'

import { errorMessage } from '@/common/api/http'
import { ErrorAlert } from '@/common/components/error-alert'
import { CartBar, CategoryFilter, PartCard, PartCardSkeleton } from '@/modules/catalog'

import { useCatalogPage } from './use-catalog-page'

const SKELETONS = [0, 1, 2] as const

export const CatalogPage = () => {
  const { t } = useTranslation('catalog')
  const {
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
    isLoading,
    isFiltered,
  } = useCatalogPage()

  return (
    <div className="flex flex-col gap-8 pb-28">
      <section className="flex flex-col gap-2">
        <h1 className="text-3xl font-bold tracking-tight sm:text-4xl">{t('hero.title')}</h1>
        <p className="max-w-2xl text-muted">
          {t('hero.text')}
        </p>
      </section>

      <section className="flex flex-col gap-4" aria-label={t('search.region')}>
        <div className="relative max-w-md">
          <Magnifier
            className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted"
            aria-hidden="true"
          />
          <Input
            aria-label={t('search.label')}
            placeholder={t('search.placeholder')}
            value={query}
            onChange={(event) => setQuery(event.target.value)}
            className="w-full pl-9"
          />
        </div>
        <CategoryFilter value={category} onChange={setCategory} counts={counts} total={parts?.length} />
      </section>

      <ErrorAlert message={loadError ? errorMessage(loadError, t('errors.load')) : null} />
      <ErrorAlert message={createError ? errorMessage(createError, t('errors.createOrder')) : null} />

      {isLoading ? (
        <ul className="grid gap-5 sm:grid-cols-2 lg:grid-cols-3" aria-label={t('state.loading')}>
          {SKELETONS.map((key) => (
            <li key={key}>
              <PartCardSkeleton />
            </li>
          ))}
        </ul>
      ) : null}

      {parts !== undefined && visible.length === 0 ? (
        <div className="rounded-2xl border border-dashed border-border p-10 text-center">
          <p className="font-medium">{isFiltered ? t('state.notFound') : t('state.empty')}</p>
          {isFiltered ? (
            <>
              <p className="mt-1 text-sm text-muted">{t('state.hint')}</p>
              <Button
                className="mt-4"
                variant="outline"
                onPress={resetFilters}
              >
                {t('state.reset')}
              </Button>
            </>
          ) : null}
        </div>
      ) : null}

      <ul className="grid gap-5 sm:grid-cols-2 lg:grid-cols-3">
        {visible.map((part, index) => (
          <li key={part.uuid}>
            <PartCard part={part} selected={selected.has(part.uuid)} onToggle={toggle} index={index} />
          </li>
        ))}
      </ul>

      {selected.size > 0 ? (
        <CartBar count={selected.size} total={total} isPending={isCreating} onClear={clear} onCheckout={checkout} />
      ) : null}
    </div>
  )
}
