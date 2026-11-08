import { act } from '@testing-library/react'

import { formatPrice } from '../lib/format'
import { i18n } from './i18n'
import { useLanguageStore } from './language.store'

describe('смена языка', () => {
  it('переводит строки и обновляет lang документа', async () => {
    expect(i18n.t('nav.catalog')).toBe('Каталог')

    await act(async () => useLanguageStore.getState().setLanguage('en'))

    expect(i18n.t('nav.catalog')).toBe('Catalog')
    expect(document.documentElement.lang).toBe('en')
  })

  it('формат чисел следует языку', async () => {
    expect(formatPrice(1234.5)).toBe('1 234,50')
    await act(async () => useLanguageStore.getState().setLanguage('en'))
    expect(formatPrice(1234.5)).toBe('1,234.50')
  })

  it('неизвестный язык в хранилище игнорируется', async () => {
    localStorage.setItem('rf:language:v1', JSON.stringify({ state: { language: 'xx' }, version: 1 }))
    await useLanguageStore.persist.rehydrate()
    expect(useLanguageStore.getState().language).not.toBe('xx')
  })
})
