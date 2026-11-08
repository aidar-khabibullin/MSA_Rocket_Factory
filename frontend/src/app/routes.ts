/**
 * Загрузчики страниц в одном месте: роутер подгружает их лениво (отдельные чанки, правило
 * bundle-dynamic-imports), а шапка предзагружает по наведению и фокусу (правило bundle-preload).
 */
export const pageLoaders = {
  catalog: () => import('@/features/catalog/CatalogPage'),
  orders: () => import('@/features/orders/OrdersPage'),
  order: () => import('@/features/orders/OrderPage'),
  auth: () => import('@/features/auth/AuthPage'),
} as const

export type PageName = keyof typeof pageLoaders

export function preloadPage(name: PageName): void {
  void pageLoaders[name]()
}
