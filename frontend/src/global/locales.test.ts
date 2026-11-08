import authEn from '@/modules/auth/locales/en.json'
import authRu from '@/modules/auth/locales/ru.json'
import catalogEn from '@/modules/catalog/locales/en.json'
import catalogRu from '@/modules/catalog/locales/ru.json'
import ordersEn from '@/modules/orders/locales/en.json'
import ordersRu from '@/modules/orders/locales/ru.json'

import commonEn from '@/common/i18n/locales/en/common.json'
import commonRu from '@/common/i18n/locales/ru/common.json'

/** Плоский список ключей вида `a.b.c`: по нему сравниваем структуру словарей. */
function keysOf(value: unknown, prefix = ''): string[] {
  if (typeof value !== 'object' || value === null) return [prefix]
  return Object.entries(value).flatMap(([key, child]) => keysOf(child, prefix ? `${prefix}.${key}` : key))
}

/** Плейсхолдеры {{name}} должны совпадать в обоих языках, иначе в одном из них значение потеряется. */
function placeholders(value: string): string[] {
  return [...value.matchAll(/{{\s*(\w+)\s*}}/g)].map((match) => match[1] ?? '').toSorted()
}

function leaves(value: unknown, prefix = ''): Array<[string, string]> {
  if (typeof value === 'string') return [[prefix, value]]
  if (typeof value !== 'object' || value === null) return []
  return Object.entries(value).flatMap(([key, child]) => leaves(child, prefix ? `${prefix}.${key}` : key))
}

const DICTIONARIES = {
  common: [commonRu, commonEn],
  auth: [authRu, authEn],
  catalog: [catalogRu, catalogEn],
  orders: [ordersRu, ordersEn],
} as const

describe('словари', () => {
  it.each(Object.entries(DICTIONARIES))('%s: наборы ключей ru и en совпадают', (_name, [ru, en]) => {
    expect(keysOf(en).toSorted()).toEqual(keysOf(ru).toSorted())
  })

  it.each(Object.entries(DICTIONARIES))('%s: плейсхолдеры в ru и en совпадают', (_name, [ru, en]) => {
    const english = new Map(leaves(en))
    for (const [key, text] of leaves(ru)) {
      // Если ключа нет в en, сравнение с пустым списком плейсхолдеров не пройдёт.
      expect({ key, placeholders: placeholders(english.get(key) ?? '') }).toEqual({ key, placeholders: placeholders(text) })
    }
  })
})
