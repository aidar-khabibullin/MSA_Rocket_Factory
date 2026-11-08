import '@testing-library/jest-dom/vitest'
import { cleanup } from '@testing-library/react'
import { afterEach, beforeEach, vi } from 'vitest'

import { i18n, useLanguageStore } from '@/common/i18n'
import '@/modules/auth/locales'
import '@/modules/catalog/locales'
import '@/modules/orders/locales'
import '@/modules/system/locales'

/**
 * Node 25 подставляет в globalThis собственный экспериментальный localStorage без методов,
 * и он перекрывает jsdom. Тесты получают предсказуемое in-memory хранилище с настоящим интерфейсом Storage.
 */
class MemoryStorage implements Storage {
  private data = new Map<string, string>()

  get length(): number {
    return this.data.size
  }
  clear(): void {
    this.data.clear()
  }
  getItem(key: string): string | null {
    return this.data.get(key) ?? null
  }
  key(index: number): string | null {
    return [...this.data.keys()][index] ?? null
  }
  removeItem(key: string): void {
    this.data.delete(key)
  }
  setItem(key: string, value: string): void {
    this.data.set(key, String(value))
  }
}

// В beforeEach, а не один раз: тесты вызывают vi.unstubAllGlobals(), и подмена пропала бы.
beforeEach(async () => {
  // jsdom отдаёт navigator.language = en-US: фиксируем русский, на нём написаны ожидания тестов.
  // Язык ставим до подмены localStorage: persist пишет в него при каждом изменении, а тесты ждут пустое хранилище.
  await i18n.changeLanguage('ru')
  useLanguageStore.setState({ language: 'ru' })
  vi.stubGlobal('localStorage', new MemoryStorage())
})

afterEach(() => {
  cleanup()
})
