import { http } from '@/common/api/http'

import type { ServiceStatus } from '../model/service-status'

export const systemApi = {
  health: (signal?: AbortSignal) => http<{ services: ServiceStatus[] }>('/api/health', { signal }),
}
