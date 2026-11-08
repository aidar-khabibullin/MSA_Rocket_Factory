import type { Part } from './part'

import { countByCategory, filterParts, LOW_STOCK_THRESHOLD, stockLevel, sumSelected } from './filters'

const part = (overrides: Partial<Part>): Part => ({
  uuid: 'id',
  name: 'Деталь',
  description: 'Описание',
  price: 100,
  stockQuantity: 10,
  category: 'CATEGORY_ENGINE',
  manufacturer: 'ACME',
  tags: [],
  ...overrides,
})

const parts = [
  part({ uuid: 'a', name: 'Двигатель RD-180', description: 'Тяжёлый носитель', price: 1000 }),
  part({ uuid: 'b', name: 'Бак Falcon', description: 'Топливный бак', category: 'CATEGORY_FUEL', price: 250.5 }),
  part({ uuid: 'c', name: 'Крыло Delta', description: 'Возвращаемая ступень', category: 'CATEGORY_WING', price: 50 }),
]

describe('filterParts', () => {
  it('без условий возвращает всё', () => {
    expect(filterParts(parts, { query: '', category: null })).toHaveLength(3)
  })

  it('ищет по названию и описанию без учёта регистра и пробелов', () => {
    expect(filterParts(parts, { query: '  falcon ', category: null }).map((p) => p.uuid)).toEqual(['b'])
    expect(filterParts(parts, { query: 'носитель', category: null }).map((p) => p.uuid)).toEqual(['a'])
  })

  it('фильтрует по категории', () => {
    expect(filterParts(parts, { query: '', category: 'CATEGORY_WING' }).map((p) => p.uuid)).toEqual(['c'])
  })

  it('условия объединяются через И', () => {
    expect(filterParts(parts, { query: 'крыло', category: 'CATEGORY_FUEL' })).toEqual([])
  })
})

describe('countByCategory', () => {
  it('считает детали по категориям', () => {
    const counts = countByCategory([...parts, part({ uuid: 'd' })])

    expect(counts.get('CATEGORY_ENGINE')).toBe(2)
    expect(counts.get('CATEGORY_FUEL')).toBe(1)
    expect(counts.get('CATEGORY_PORTHOLE')).toBeUndefined()
  })
})

describe('sumSelected', () => {
  it('суммирует только выбранные детали', () => {
    expect(sumSelected(parts, new Set(['a', 'b']))).toBeCloseTo(1250.5)
    expect(sumSelected(parts, new Set())).toBe(0)
  })
})

describe('stockLevel', () => {
  it.each([
    [0, 'out'],
    [-1, 'out'],
    [1, 'low'],
    [LOW_STOCK_THRESHOLD, 'low'],
    [LOW_STOCK_THRESHOLD + 1, 'ok'],
  ])('остаток %i -> %s', (quantity, level) => {
    expect(stockLevel(quantity)).toBe(level)
  })
})
