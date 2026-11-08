/** Ключи TanStack Query: по ним данные кэшируются, дедуплицируются и инвалидируются. */
export const catalogKeys = {
  parts: (sessionUuid: string) => ['parts', sessionUuid] as const,
}
