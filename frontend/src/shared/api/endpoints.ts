import { http } from './http'
import type { Order, Part, PaymentMethod, ServiceStatus } from './types'

export const authApi = {
  register: (login: string, email: string, password: string) =>
    http<{ user_uuid: string }>('/api/auth/register', { method: 'POST', body: { login, email, password } }),

  login: (login: string, password: string) =>
    http<{ session_uuid: string }>('/api/auth/login', { method: 'POST', body: { login, password } }),

  me: (sessionUuid: string) =>
    http<{ user_uuid: string; login: string; email: string }>('/api/auth/me', { sessionUuid }),
}

export const catalogApi = {
  parts: (sessionUuid: string, signal?: AbortSignal) => http<Part[]>('/api/parts', { sessionUuid, signal }),
}

export const ordersApi = {
  create: (sessionUuid: string, userUuid: string, partUuids: string[]) =>
    http<{ order_uuid: string; total_price?: number }>('/api/v1/orders', {
      method: 'POST',
      sessionUuid,
      body: { user_uuid: userUuid, part_uuids: partUuids },
    }),

  get: (sessionUuid: string, orderUuid: string, signal?: AbortSignal) =>
    http<Order>(`/api/v1/orders/${orderUuid}`, { sessionUuid, signal }),

  pay: (sessionUuid: string, orderUuid: string, paymentMethod: PaymentMethod) =>
    http<{ transaction_uuid: string }>(`/api/v1/orders/${orderUuid}/pay`, {
      method: 'POST',
      sessionUuid,
      body: { payment_method: paymentMethod },
    }),

  cancel: (sessionUuid: string, orderUuid: string) =>
    http<void>(`/api/v1/orders/${orderUuid}/cancel`, { method: 'POST', sessionUuid }),
}

export const systemApi = {
  health: (signal?: AbortSignal) => http<{ services: ServiceStatus[] }>('/api/health', { signal }),
}
