/** Ключи SWR в одном месте: по ним данные кэшируются, дедуплицируются и инвалидируются. */
export const swrKeys = {
  parts: (sessionUuid: string) => ['parts', sessionUuid] as const,
  order: (sessionUuid: string, orderUuid: string) => ['order', sessionUuid, orderUuid] as const,
  orders: (sessionUuid: string, orderUuids: readonly string[]) => ['orders', sessionUuid, ...orderUuids] as const,
  health: ['health'] as const,
}
