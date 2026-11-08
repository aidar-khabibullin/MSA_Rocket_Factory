import { useTranslation } from 'react-i18next'

import { CATEGORY_KEYS, type CategoryKey } from '../model/categories'
import { CategoryChip } from './category-chip'

interface CategoryFilterProps {
  value: CategoryKey | null
  onChange: (category: CategoryKey | null) => void
  counts: ReadonlyMap<string, number>
  total: number | undefined
}

export const CategoryFilter = ({ value, onChange, counts, total }: CategoryFilterProps) => {
  const { t } = useTranslation('catalog')
  return (
    <fieldset className="flex min-w-0 flex-wrap gap-2 border-0 p-0">
      <legend className="sr-only">{t('categories.legend')}</legend>
      <CategoryChip active={value === null} onClick={() => onChange(null)}>
        {t('categories.all')}
        {total === undefined ? '' : ` · ${total}`}
      </CategoryChip>
      {CATEGORY_KEYS.map((key) => (
        <CategoryChip key={key} active={value === key} onClick={() => onChange(value === key ? null : key)}>
          {t(`categories.${key}`)} · {counts.get(key) ?? 0}
        </CategoryChip>
      ))}
    </fieldset>
  )
}
