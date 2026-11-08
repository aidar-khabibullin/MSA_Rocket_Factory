import '@testing-library/jest-dom/vitest'
import { cleanup } from '@testing-library/react'
import { afterEach, beforeEach, vi } from 'vitest'

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
beforeEach(() => {
  vi.stubGlobal('localStorage', new MemoryStorage())
})

afterEach(() => {
  cleanup()
})
