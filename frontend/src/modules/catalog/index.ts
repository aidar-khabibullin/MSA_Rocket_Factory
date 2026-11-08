import './locales'

// Публичный API модуля: снаружи импортируем только отсюда, а не из внутренних папок.
export { CartBar } from './components/cart-bar'
export { CategoryFilter } from './components/category-filter'
export { PartCard, PartCardSkeleton } from './components/part-card'
export { usePartsIndex, useParts } from './hooks/use-parts'
export {
  categoryFromParam,
  categoryLabelKey,
  categoryMeta,
  categoryToParam,
  type CategoryKey,
} from './model/categories'
export { countByCategory, filterParts, sumSelected } from './model/filters'
export type { Part } from './model/part'
export { useCartStore } from './store/cart.store'
