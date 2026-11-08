import { groupPartUuids } from './OrderComposition'

describe('groupPartUuids', () => {
  it('схлопывает повторы и сохраняет порядок первого появления', () => {
    expect(groupPartUuids(['a', 'b', 'a', 'a', 'c'])).toEqual([
      { id: 'a', quantity: 3 },
      { id: 'b', quantity: 1 },
      { id: 'c', quantity: 1 },
    ])
  })

  it('пустой список даёт пустой результат', () => {
    expect(groupPartUuids([])).toEqual([])
  })
})
