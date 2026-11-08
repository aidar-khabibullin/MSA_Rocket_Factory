import { act, renderHook } from '@testing-library/react'

import { STORAGE_KEYS } from '@/common/lib/storage'

import { getSession, parseSession, signIn, signOut, useSession, useSessionStore } from './session.store'

const session = { sessionUuid: 's-1', userUuid: 'u-1', login: 'demo' }

describe('session store', () => {
  beforeEach(() => useSessionStore.setState({ session: null }))

  it('вход сохраняет сессию в состоянии и в localStorage, выход очищает', () => {
    signIn(session)

    expect(getSession()).toEqual(session)
    expect(JSON.parse(localStorage.getItem(STORAGE_KEYS.session) ?? '').state.session).toEqual(session)

    signOut()

    expect(getSession()).toBeNull()
    expect(JSON.parse(localStorage.getItem(STORAGE_KEYS.session) ?? '').state.session).toBeNull()
  })

  it('компонент перерисовывается при входе и выходе', () => {
    const { result } = renderHook(() => useSession())
    expect(result.current).toBeNull()

    act(() => signIn(session))
    expect(result.current).toEqual(session)

    act(() => signOut())
    expect(result.current).toBeNull()
  })

  it('восстанавливает сессию из хранилища', async () => {
    localStorage.setItem(STORAGE_KEYS.session, JSON.stringify({ state: { session }, version: 1 }))

    await useSessionStore.persist.rehydrate()

    expect(getSession()).toEqual(session)
  })

  it('игнорирует повреждённую сессию в хранилище', async () => {
    localStorage.setItem(STORAGE_KEYS.session, JSON.stringify({ state: { session: { sessionUuid: 1 } }, version: 1 }))

    await useSessionStore.persist.rehydrate()

    expect(getSession()).toBeNull()
  })
})

describe('parseSession', () => {
  it.each([null, undefined, 'строка', 42, [], {}, { sessionUuid: 's', userUuid: 'u' }])(
    'отклоняет %j',
    (value) => {
      expect(parseSession(value)).toBeNull()
    },
  )

  it('принимает объект ожидаемой формы и отбрасывает лишние поля', () => {
    expect(parseSession({ ...session, extra: 'x' })).toEqual(session)
  })
})
