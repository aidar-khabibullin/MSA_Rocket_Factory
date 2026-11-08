import { http } from '@/common/api/http'

import type { Part } from '../model/part'
import { partsFromDto } from './part.adapter'
import type { ListPartsResponseDto } from './part.dto'

export const catalogApi = {
  parts: async (sessionUuid: string, signal?: AbortSignal): Promise<Part[]> =>
    partsFromDto(await http<ListPartsResponseDto>('/api/parts', { sessionUuid, signal })),
}
