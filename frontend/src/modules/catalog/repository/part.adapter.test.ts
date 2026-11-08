import { partFromDto } from './part.adapter'

describe('partFromDto', () => {
  it('переводит поля бэкенда в camelCase-модель', () => {
    expect(
      partFromDto({
        part_uuid: 'p1',
        name: 'Двигатель',
        description: 'Тяга',
        price: 100,
        stock_quantity: 3,
        category: 'CATEGORY_ENGINE',
        manufacturer: 'ACME',
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

  it('подставляет пустой список, если бэкенд не прислал теги', () => {
    const part = partFromDto({
      part_uuid: 'p1',
      name: 'n',
      description: 'd',
      price: 1,
      stock_quantity: 0,
      category: 'c',
      manufacturer: 'm',
    })
    expect(part.tags).toEqual([])
  })
})
