/** Ключи SWR: по ним данные кэшируются, дедуплицируются и инвалидируются. */
export const ordersKeys = {
  order: (sessionUuid: string, orderUuid: string) => ['order', sessionUuid, orderUuid] as const,
  orders: (sessionUuid: string, orderUuids: readonly string[]) => ['orders', sessionUuid, ...orderUuids] as const,
}
