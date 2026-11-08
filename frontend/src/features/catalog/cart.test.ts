import { signIn, signOut, useSessionStore } from '@/features/auth/session'

import { useCartStore } from './hooks'

describe('cart store', () => {
  beforeEach(() => {
    useSessionStore.setState({ session: null })
    useCartStore.getState().clear()
  })

  it('toggle добавляет и убирает деталь', () => {
    const { toggle } = useCartStore.getState()

    toggle('a')
    toggle('b')
    expect([...useCartStore.getState().selected]).toEqual(['a', 'b'])

    toggle('a')
    expect([...useCartStore.getState().selected]).toEqual(['b'])
  })

  it('каждое изменение создаёт новый Set, иначе подписчики не заметят обновление', () => {
    const before = useCartStore.getState().selected

    useCartStore.getState().toggle('a')

    expect(useCartStore.getState().selected).not.toBe(before)
  })

  it('действия стабильны: ссылки на них не меняются между обновлениями', () => {
    const { toggle } = useCartStore.getState()

    toggle('a')

    expect(useCartStore.getState().toggle).toBe(toggle)
  })

  it('корзина очищается при выходе, чтобы не достаться следующему пользователю', () => {
    signIn({ sessionUuid: 's', userUuid: 'u', login: 'demo' })
    useCartStore.getState().toggle('a')

    signOut()

    expect(useCartStore.getState().selected.size).toBe(0)
  })
})
