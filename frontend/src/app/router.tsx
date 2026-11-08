import { createBrowserRouter, redirect } from 'react-router-dom'

import { getSession } from '@/modules/auth'

import { Layout } from '@/app/layout'
import { pageLoaders } from '@/app/routes'
import { RouteError } from '@/app/route-error'

/** Гость уходит на /login до рендера защищённых страниц. */
function requireAuth() {
  if (!getSession()) throw redirect('/login')
  return null
}

export const router = createBrowserRouter([
  {
    path: '/login',
    errorElement: <RouteError />,
    lazy: async () => ({ Component: (await pageLoaders.auth()).AuthPage }),
  },
  {
    path: '/',
    Component: Layout,
    loader: requireAuth,
    errorElement: <RouteError />,
    // Сбой страницы (в том числе загрузки её чанка) показывается внутри раскладки: шапка и навигация остаются рабочими.
    children: [
      {
        errorElement: <RouteError inline />,
        children: [
          { index: true, lazy: async () => ({ Component: (await pageLoaders.catalog()).CatalogPage }) },
          { path: 'orders', lazy: async () => ({ Component: (await pageLoaders.orders()).OrdersPage }) },
          { path: 'orders/:orderId', lazy: async () => ({ Component: (await pageLoaders.order()).OrderPage }) },
          { path: '*', loader: () => redirect('/') },
        ],
      },
    ],
  },
])
