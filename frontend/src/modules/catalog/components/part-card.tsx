import Check from '@gravity-ui/icons/Check'
import { Button } from '@heroui/react/button'
import { Chip } from '@heroui/react/chip'
import { memo } from 'react'
import { useTranslation } from 'react-i18next'

import type { Part } from '../model/part'
import { formatPrice } from '@/common/lib/format'

import { categoryLabelKey, categoryMeta } from '../model/categories'
import { stockLevel, type StockLevel } from '../model/filters'

interface PartCardProps {
  part: Part
  selected: boolean
  onToggle: (partUuid: string) => void
  index: number
}

const STOCK_TONE: Record<StockLevel, string> = {
  out: 'text-danger',
  low: 'text-warning',
  ok: 'text-muted',
}

/** memo: при вводе в поиск или выборе другой детали карточка с теми же props не перерисовывается. */
export const PartCard = memo(function PartCard({ part, selected, onToggle, index }: PartCardProps) {
  const { t } = useTranslation('catalog')
  const { icon: Icon } = categoryMeta(part.category)
  const stock = stockLevel(part.stockQuantity)

  return (
    <article
      className={`lift rise flex h-full flex-col gap-4 rounded-2xl border bg-surface p-5 ${
        selected
          ? 'border-accent shadow-[0_0_0_1px_var(--accent),0_8px_32px_-12px_var(--accent)]'
          : 'border-border hover:border-accent/50'
      }`}
      style={{ animationDelay: `${Math.min(index, 8) * 40}ms` }}
    >
      <div className="flex items-start justify-between gap-3">
        <span className="grid size-11 place-items-center rounded-xl bg-accent/12 text-accent" aria-hidden="true">
          <Icon className="size-6" />
        </span>
        <Chip size="sm" variant="soft">
          {t(categoryLabelKey(part.category))}
        </Chip>
      </div>

      <div className="flex flex-1 flex-col gap-1">
        <h3 className="text-base font-bold leading-snug">{part.name}</h3>
        <p className="text-sm leading-relaxed text-muted">{part.description}</p>
        <p className="mt-1 text-xs text-muted">{t('card.manufacturer', { name: part.manufacturer })}</p>
      </div>

      <div className="flex items-end justify-between gap-3 border-t border-border/60 pt-4">
        <div>
          <p className="text-xl font-bold tabular-nums">{formatPrice(part.price)}</p>
          <p className={`text-xs ${STOCK_TONE[stock]}`}>{t(`stock.${stock}`, { quantity: part.stockQuantity })}</p>
        </div>
        <Button
          variant={selected ? 'secondary' : 'primary'}
          isDisabled={stock === 'out' && !selected}
          onPress={() => onToggle(part.uuid)}
          aria-pressed={selected}
          aria-label={t(selected ? 'card.removeAria' : 'card.addAria', { name: part.name })}
        >
          {selected ? (
            <>
              <Check /> {t('card.inOrder')}
            </>
          ) : (
            t('card.add')
          )}
        </Button>
      </div>
    </article>
  )
})

export const PartCardSkeleton = () => {
  return (
    <div
      className="flex h-64 animate-pulse flex-col gap-4 rounded-2xl border border-border bg-surface p-5"
      aria-hidden="true"
    >
      <div className="flex justify-between">
        <div className="size-11 rounded-xl bg-border" />
        <div className="h-6 w-20 rounded-full bg-border" />
      </div>
      <div className="h-5 w-3/4 rounded bg-border" />
      <div className="h-4 w-full rounded bg-border" />
      <div className="mt-auto h-8 w-1/2 rounded bg-border" />
    </div>
  )
}
