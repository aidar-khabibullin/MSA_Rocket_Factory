import useSWR from 'swr'

import type { ServiceStatus } from '../model/service-status'
import { systemApi } from '../repository/system.api'
import { systemKeys } from '../repository/system.keys'

const POLL_INTERVAL_MS = 30_000

export interface ServiceHealth {
  /** undefined — ещё не проверяли; пустой массив недоступен: при ошибке шлюза isGatewayDown = true. */
  services: ServiceStatus[] | undefined
  checkedAt: Date | undefined
  isGatewayDown: boolean
  isDegraded: boolean
  isChecking: boolean
  refresh: () => void
}

/**
 * Опрос раз в 30 секунд, пока вкладка видна. Баннер и подвал вызывают хук независимо,
 * но SWR дедуплицирует запросы: на сервер уходит один.
 */
export function useServiceStatus(): ServiceHealth {
  const { data, error, isValidating, mutate } = useSWR(
    systemKeys.health,
    async () => ({ ...(await systemApi.health()), checkedAt: new Date() }),
    { refreshInterval: POLL_INTERVAL_MS, shouldRetryOnError: false },
  )

  const isGatewayDown = error !== undefined
  const services = data?.services

  return {
    services,
    checkedAt: data?.checkedAt,
    isGatewayDown,
    isDegraded: isGatewayDown || (services?.some((service) => !service.ok) ?? false),
    isChecking: isValidating,
    refresh: () => void mutate(),
  }
}
