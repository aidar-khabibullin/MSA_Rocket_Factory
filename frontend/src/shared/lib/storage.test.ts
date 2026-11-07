import { migrateLegacyStorage, STORAGE_KEYS } from './storage'

describe('migrateLegacyStorage', () => {
  it('переносит сессию, тему и заказы старой версии в формат persist и удаляет старые ключи', () => {
    localStorage.setItem('rf.session', JSON.stringify({ sessionUuid: 's', userUuid: 'u', login: 'demo' }))
    localStorage.setItem('rf.theme', 'light')
    localStorage.setItem('rf.orders.user-1', JSON.stringify(['o1', 'o2']))

    migrateLegacyStorage()

    expect(JSON.parse(localStorage.getItem(STORAGE_KEYS.session) ?? '')).toEqual({
      state: { session: { sessionUuid: 's', userUuid: 'u', login: 'demo' } },
      version: 1,
    })
    expect(JSON.parse(localStorage.getItem(STORAGE_KEYS.theme) ?? '')).toEqual({
      state: { theme: 'light' },
      version: 1,
    })
    expect(JSON.parse(localStorage.getItem(STORAGE_KEYS.orders) ?? '')).toEqual({
      state: { byUser: { 'user-1': ['o1', 'o2'] } },
      version: 1,
    })
    expect(localStorage.getItem('rf.session')).toBeNull()
    expect(localStorage.getItem('rf.theme')).toBeNull()
    expect(localStorage.getItem('rf.orders.user-1')).toBeNull()
  })

  it('не затирает данные новой версии', () => {
    const current = JSON.stringify({ state: { theme: 'dark' }, version: 1 })
    localStorage.setItem(STORAGE_KEYS.theme, current)
    localStorage.setItem('rf.theme', 'light')

    migrateLegacyStorage()

    expect(localStorage.getItem(STORAGE_KEYS.theme)).toBe(current)
  })

  it('не падает на повреждённых данных', () => {
    localStorage.setItem('rf.session', '{not json')

    expect(() => migrateLegacyStorage()).not.toThrow()
  })

  it('ничего не делает без старых ключей', () => {
    migrateLegacyStorage()

    expect(localStorage.length).toBe(0)
  })
})
