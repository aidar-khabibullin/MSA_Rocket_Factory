import { ToastProvider } from '@heroui/react/toast'
import { RouterProvider } from 'react-router-dom'
import { SWRConfig } from 'swr'

import { ApiError } from '@/common/api/http'
import { i18n } from '@/common/i18n'
import { toastService } from '@/common/service/toast-service'
import { signOut } from '@/modules/auth'

import { router } from './router'

export function App() {
  return (
    <SWRConfig
      value={{
        // Сессия истекла или отозвана: выходим, а не показываем «Ошибка 401» на каждой странице.
        onError: (error) => {
          if (error instanceof ApiError && error.status === 401) {
            signOut()
            toastService.showWarning(i18n.t('toast.sessionExpired'))
          }
        },
        shouldRetryOnError: false,
      }}
    >
      <RouterProvider router={router} />
      <ToastProvider />
    </SWRConfig>
  )
}
