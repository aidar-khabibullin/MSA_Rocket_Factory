import { partFromDto, partsFromDto } from './part.adapter'

describe('partFromDto', () => {
  it('переводит поля бэкенда в camelCase-модель: int64 строкой, вложенный производитель', () => {
    expect(
      partFromDto({
        part_uuid: 'p1',
        name: 'Двигатель',
        description: 'Тяга',
        price: 100,
        stock_quantity: '3',
        category: 'CATEGORY_ENGINE',
        manufacturer: { name: 'ACME', country: 'USA', url: 'https://acme.test' },
        tags: ['heavy'],
      }),
    ).toEqual({
      uuid: 'p1',
      name: 'Двигатель',
      description: 'Тяга',
      price: 100,
      stockQuantity: 3,
      category: 'CATEGORY_ENGINE',
      manufacturer: 'ACME',
      tags: ['heavy'],
    })
  })

  it('подставляет пустой список тегов и пустого производителя, если бэкенд их не прислал', () => {
    const part = partFromDto({
      part_uuid: 'p1',
      name: 'n',
      description: 'd',
      price: 1,
      stock_quantity: '0',
      category: 'c',
    })
    expect(part.tags).toEqual([])
    expect(part.manufacturer).toBe('')
  })

  it('большой остаток из строки int64 остаётся числом', () => {
    const part = partFromDto({
      part_uuid: 'p1',
      name: 'n',
      description: 'd',
      price: 1,
      stock_quantity: '1000000',
      category: 'c',
    })
    expect(part.stockQuantity).toBe(1_000_000)
  })

  it('нечисловой остаток даёт 0, а не NaN', () => {
    const part = partFromDto({ part_uuid: 'p1', name: 'n', description: 'd', price: 1, stock_quantity: 'abc', category: 'c' })
    expect(part.stockQuantity).toBe(0)
  })
})

describe('partsFromDto', () => {
  it('достаёт детали из обёртки и возвращает пустой список без поля parts', () => {
    expect(partsFromDto({})).toEqual([])
    expect(
      partsFromDto({ parts: [{ part_uuid: 'p1', name: 'n', description: 'd', price: 1, stock_quantity: '2', category: 'c' }] }),
    ).toHaveLength(1)
  })
})
