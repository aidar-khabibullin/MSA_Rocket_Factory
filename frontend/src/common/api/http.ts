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

function parseJSON(text: string): unknown {
  try {
    return JSON.parse(text)
  } catch {
    return undefined
  }
}

export async function http<T>(path: string, options: RequestOptions = {}): Promise<T> {
  const { method = 'GET', body, sessionUuid, signal } = options

  const headers = new Headers()
  if (body !== undefined) headers.set('Content-Type', 'application/json')
  if (sessionUuid) headers.set('X-Session-Uuid', sessionUuid)

  let response: Response
  try {
    response = await fetch(path, {
      method,
      headers,
      signal,
      body: body === undefined ? undefined : JSON.stringify(body),
    })
  } catch (error) {
    if (error instanceof DOMException && error.name === 'AbortError') throw error
    throw new ApiError(NETWORK_ERROR, i18n.t('errors.network'))
  }

  const text = response.status === 204 ? '' : await response.text()
  const payload = text ? parseJSON(text) : undefined

  if (!response.ok) {
    throw new ApiError(response.status, extractMessage(payload) ?? i18n.t('errors.status', { status: response.status }))
  }
  return payload as T
}

/** Сообщение для пользователя из любой ошибки. */
export function errorMessage(error: unknown, fallback: string): string {
  return error instanceof ApiError ? error.message : fallback
}
