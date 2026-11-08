import { STORAGE_KEYS } from '@/common/lib/storage'

import { rememberOrder, useOrderHistoryStore } from './order-history.store'

const ids = (user: string) => useOrderHistoryStore.getState().byUser[user]

describe('order history', () => {
  beforeEach(() => useOrderHistoryStore.setState({ byUser: {} }))

  it('новые заказы идут в начало списка', () => {
    rememberOrder('u1', 'o1')
    rememberOrder('u1', 'o2')

    expect(ids('u1')).toEqual(['o2', 'o1'])
  })

  it('не дублирует уже запомненный заказ и не меняет состояние', () => {
    rememberOrder('u1', 'o1')
    const before = useOrderHistoryStore.getState().byUser

    rememberOrder('u1', 'o1')

    expect(useOrderHistoryStore.getState().byUser).toBe(before)
  })

  it('история разных пользователей не пересекается', () => {
    rememberOrder('u1', 'o1')
    rememberOrder('u2', 'o2')

    expect(ids('u1')).toEqual(['o1'])
    expect(ids('u2')).toEqual(['o2'])
  })

  it('сохраняет только идентификаторы', () => {
    rememberOrder('u1', 'o1')

    expect(JSON.parse(localStorage.getItem(STORAGE_KEYS.orders) ?? '')).toEqual({
      state: { byUser: { u1: ['o1'] } },
      version: 1,
    })
  })

  it('игнорирует повреждённые данные в хранилище', async () => {
    localStorage.setItem(STORAGE_KEYS.orders, JSON.stringify({ state: { byUser: { u1: [1, 2] } }, version: 1 }))

    await useOrderHistoryStore.persist.rehydrate()

    expect(useOrderHistoryStore.getState().byUser).toEqual({})
  })
})
