import Cpu from '@gravity-ui/icons/Cpu'
import Droplet from '@gravity-ui/icons/Droplet'
import Flame from '@gravity-ui/icons/Flame'
import Rocket from '@gravity-ui/icons/Rocket'
import Thunderbolt from '@gravity-ui/icons/Thunderbolt'
import type { ComponentType, SVGProps } from 'react'

export type CategoryIcon = ComponentType<SVGProps<SVGSVGElement>>

interface CategoryMeta {
  icon: CategoryIcon
}

/** Порядок фильтров и ключи совпадают со значениями Category в inventory.proto. */
export const CATEGORIES = {
  CATEGORY_ENGINE: { icon: Flame },
  CATEGORY_FUEL: { icon: Droplet },
  CATEGORY_PORTHOLE: { icon: Cpu },
  CATEGORY_WING: { icon: Rocket },
} as const satisfies Record<string, CategoryMeta>

export type CategoryKey = keyof typeof CATEGORIES

export const CATEGORY_KEYS = Object.keys(CATEGORIES) as CategoryKey[]

const FALLBACK: CategoryMeta = { icon: Thunderbolt }

export function categoryMeta(key: string): CategoryMeta {
  return key in CATEGORIES ? CATEGORIES[key as CategoryKey] : FALLBACK
}

/** Ключ подписи категории в словаре catalog; неизвестная категория сервера попадает в «Прочее». */
export const categoryLabelKey = (key: string): `categories.${CategoryKey}` | 'categories.other' =>
  key in CATEGORIES ? `categories.${key as CategoryKey}` : 'categories.other'
