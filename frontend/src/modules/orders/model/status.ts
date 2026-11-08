import type { OrderStatus } from './order'

export type StatusTone = 'success' | 'accent' | 'danger' | 'warning'

export const STATUS_TONE: Record<OrderStatus, StatusTone> = {
  PENDING_PAYMENT: 'warning',
  PAID: 'accent',
  ASSEMBLED: 'success',
  CANCELLED: 'danger',
}

/** Шаги прогресса заказа (подписи — ключи steps.* в словаре); отменённый заказ показывается отдельно. */
export const PROGRESS_STEPS = [{ id: 'created' }, { id: 'paid' }, { id: 'assembled' }] as const

const REACHED_STEP: Record<Exclude<OrderStatus, 'CANCELLED'>, number> = {
  PENDING_PAYMENT: 0,
  PAID: 1,
  ASSEMBLED: 2,
}

/** Индекс последнего достигнутого шага; для отменённого заказа шагов нет. */
export function reachedStep(status: OrderStatus): number | null {
  return status === 'CANCELLED' ? null : REACHED_STEP[status]
}

/** Сборка асинхронна (Kafka): пока заказ оплачен, статус нужно опрашивать до ASSEMBLED. */
export const isAwaitingAssembly = (status: OrderStatus | undefined): boolean => status === 'PAID'

/** Бэкенд отдаёт нулевой UUID, пока заказ не оплачен. */
export const isRealUuid = (value: string | undefined): value is string => !!value && !/^[0-]+$/.test(value)
