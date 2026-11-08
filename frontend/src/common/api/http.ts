import ky, { isHTTPError } from 'ky'

import { i18n } from '@/common/i18n'

export class ApiError extends Error {
  readonly status: number

  constructor(status: number, message: string) {
    super(message)
    this.name = 'ApiError'
    this.status = status
  }
}

/** Статус 0 — запрос не дошёл до сервера. */
export const NETWORK_ERROR = 0

interface RequestOptions {
  method?: 'GET' | 'POST'
  body?: unknown
  sessionUuid?: string
  signal?: AbortSignal
}

/**
 * Повторы отключены: POST (оплата, создание заказа) нельзя безопасно повторять вслепую, а ошибки показываем сразу.
 * baseUrl нужен, чтобы относительные пути работали и там, где Request не знает о location (jsdom в тестах).
 */
const client = ky.create({ retry: 0, timeout: false, baseUrl: globalThis.location.origin })

/** Сервисы отдают ошибки в разных формах: {message}, {error_message}, {error: {message}}. */
function extractMessage(body: unknown): string | undefined {
  if (typeof body !== 'object' || body === null) return undefined
  const record = body as Record<string, unknown>
  if (typeof record.message === 'string') return record.message
  if (typeof record.error_message === 'string') return record.error_message
  const nested = record.error
  if (typeof nested === 'object' && nested !== null) {
    const message = (nested as Record<string, unknown>).message
    if (typeof message === 'string') return message
  }
  return undefined
}

/** ky разбирает тело ошибки только при Content-Type: application/json; иначе оно приходит строкой. */
function errorBody(data: unknown): unknown {
  if (typeof data !== 'string') return data
  try {
    return JSON.parse(data)
  } catch {
    return undefined
  }
}

export async function http<T>(path: string, options: RequestOptions = {}): Promise<T> {
  const { method = 'GET', body, sessionUuid, signal } = options

  try {
    const response = await client(path, {
      method,
      json: body,
      signal,
      headers: sessionUuid ? { 'Session-Uuid': sessionUuid } : undefined,
    })
    if (response.status === 204) return undefined as T
    return (await response.json()) as T
  } catch (error) {
    if (isHTTPError(error)) {
      const { status } = error.response
      throw new ApiError(status, extractMessage(errorBody(error.data)) ?? i18n.t('errors.status', { status }))
    }
    if (error instanceof DOMException && error.name === 'AbortError') throw error
    throw new ApiError(NETWORK_ERROR, i18n.t('errors.network'))
  }
}

/** Сообщение для пользователя из любой ошибки. */
export function errorMessage(error: unknown, fallback: string): string {
  return error instanceof ApiError ? error.message : fallback
}
