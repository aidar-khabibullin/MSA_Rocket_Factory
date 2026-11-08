/**
 * Загрузчики страниц в одном месте: роутер подгружает их лениво (отдельные чанки, правило
 * bundle-dynamic-imports), а шапка предзагружает по наведению и фокусу (правило bundle-preload).
 */
export const pageLoaders = {
  catalog: () => import('@/pages/catalog-page/catalog-page'),
  orders: () => import('@/pages/orders-page/orders-page'),
  order: () => import('@/pages/order-page/order-page'),
  auth: () => import('@/pages/auth-page/auth-page'),
} as const

export type PageName = keyof typeof pageLoaders

export function preloadPage(name: PageName): void {
  void pageLoaders[name]()
}
