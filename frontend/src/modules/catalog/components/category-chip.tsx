import type { ReactNode } from 'react'

interface CategoryChipProps {
  active: boolean
  onClick: () => void
  children: ReactNode
}

export const CategoryChip = ({ active, onClick, children }: CategoryChipProps) => {
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
