import { createBrowserRouter, redirect } from 'react-router-dom'

import { getSession } from '@/features/auth/session'

import { Layout } from './Layout'
import { pageLoaders } from './routes'
import { RouteError } from './RouteError'

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
    children: [
      { index: true, lazy: async () => ({ Component: (await pageLoaders.catalog()).CatalogPage }) },
      { path: 'orders', lazy: async () => ({ Component: (await pageLoaders.orders()).OrdersPage }) },
      { path: 'orders/:orderId', lazy: async () => ({ Component: (await pageLoaders.order()).OrderPage }) },
      { path: '*', loader: () => redirect('/') },
    ],
  },
])
