import type { ReactNode } from 'react'

import { CATEGORIES, CATEGORY_KEYS, type CategoryKey } from './categories'

interface CategoryFilterProps {
  value: CategoryKey | null
  onChange: (category: CategoryKey | null) => void
  counts: ReadonlyMap<string, number>
  total: number | undefined
}

export function CategoryFilter({ value, onChange, counts, total }: CategoryFilterProps) {
  return (
    <fieldset className="flex min-w-0 flex-wrap gap-2 border-0 p-0">
      <legend className="sr-only">Категории</legend>
      <Chip active={value === null} onClick={() => onChange(null)}>
        Все{total === undefined ? '' : ` · ${total}`}
      </Chip>
      {CATEGORY_KEYS.map((key) => (
        <Chip key={key} active={value === key} onClick={() => onChange(value === key ? null : key)}>
          {CATEGORIES[key].label} · {counts.get(key) ?? 0}
        </Chip>
      ))}
    </fieldset>
  )
}

function Chip({ active, onClick, children }: { active: boolean; onClick: () => void; children: ReactNode }) {
  return (
    <button
      type="button"
      aria-pressed={active}
      onClick={onClick}
      className={`min-h-9 rounded-full border px-4 text-sm font-medium transition-colors ${
        active
          ? 'border-accent bg-accent text-accent-foreground'
          : 'border-border bg-surface text-muted hover:border-accent/60 hover:text-foreground'
      }`}
    >
      {children}
    </button>
  )
}
