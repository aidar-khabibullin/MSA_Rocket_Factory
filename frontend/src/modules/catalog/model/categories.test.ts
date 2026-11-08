import { describe, expect, it } from 'vitest'

import { CATEGORY_KEYS, categoryFromParam, categoryToParam } from './categories'

describe('категория в URL', () => {
  it('значение читается человеком: без префикса и в нижнем регистре', () => {
    expect(categoryToParam('CATEGORY_ENGINE')).toBe('engine')
  })

  it('преобразование обратимо для каждой категории', () => {
    for (const key of CATEGORY_KEYS) expect(categoryFromParam(categoryToParam(key))).toBe(key)
  })

  it('неизвестное или пустое значение даёт null', () => {
    expect(categoryFromParam('unknown')).toBeNull()
    expect(categoryFromParam('')).toBeNull()
    expect(categoryFromParam(null)).toBeNull()
  })
})
