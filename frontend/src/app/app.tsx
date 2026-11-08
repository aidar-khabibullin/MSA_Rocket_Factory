import { QueryClientProvider } from '@tanstack/react-query'
import { ToastProvider } from '@heroui/react/toast'
import { RouterProvider } from 'react-router-dom'

import { onUnauthorized, queryClient } from '@/common/api/query-client'
import { i18n } from '@/common/i18n'
import { toastService } from '@/common/service/toast-service'
import { signOut } from '@/modules/auth'

import { router } from './router'

// Сессия истекла или отозвана: выходим, а не показываем «Ошибка 401» на каждой странице.
onUnauthorized(() => {
  signOut()
  toastService.showWarning(i18n.t('toast.sessionExpired'))
})

export function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <RouterProvider router={router} />
      <ToastProvider />
    </QueryClientProvider>
  )
}
