import type { ReactNode } from 'react'

interface AsyncGateProps<T> {
  data: T | undefined
  isLoading: boolean
  /** Показывается, когда загрузка закончилась, а данных нет (ошибка запроса или пустой ответ). */
  fallback: ReactNode
  loader: ReactNode
  children: (data: T) => ReactNode
}

/**
 * Одно место для ветвления «загрузка / нет данных / данные»: страницы не повторяют
 * цепочки if, а потребитель получает `data` уже без undefined.
 */
export function AsyncGate<T>({ data, isLoading, fallback, loader, children }: AsyncGateProps<T>) {
  if (isLoading) return loader
  if (data === undefined) return fallback
  return children(data)
}
