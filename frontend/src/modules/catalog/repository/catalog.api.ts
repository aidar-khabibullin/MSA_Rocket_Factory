import { http } from '@/common/api/http'

import type { Part } from '../model/part'
import { partFromDto } from './part.adapter'
import type { PartDto } from './part.dto'

export const catalogApi = {
  parts: async (sessionUuid: string, signal?: AbortSignal): Promise<Part[]> =>
    (await http<PartDto[]>('/api/parts', { sessionUuid, signal })).map(partFromDto),
}
