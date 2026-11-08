import { useCallback } from 'react'
import { useSearchParams } from 'react-router-dom'

/**
 * Один параметр адресной строки как состояние. Запись идёт через replace: фильтры не засоряют историю
 * браузера, а ссылку с ними можно скопировать и открыть. Остальные параметры URL не затрагиваются.
 */
export const useUrlParam = (key: string): readonly [string | null, (value: string | null) => void] => {
  const [searchParams, setSearchParams] = useSearchParams()

  const set = useCallback(
    (value: string | null) => {
      setSearchParams(
        (previous) => {
          const next = new URLSearchParams(previous)
          if (value === null) next.delete(key)
          else next.set(key, value)
          return next
        },
        { replace: true },
      )
    },
    [key, setSearchParams],
  )

  return [searchParams.get(key), set]
}
