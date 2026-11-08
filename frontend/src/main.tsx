import './bootstrap' // должен быть первым: см. комментарий в bootstrap.ts

import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { RouterProvider } from 'react-router-dom'
import { SWRConfig } from 'swr'

import { router } from '@/app/router'
import { signOut } from '@/features/auth/session'
import { ApiError } from '@/shared/api/http'

import './index.css'

const container = document.getElementById('root')
if (!container) throw new Error('Корневой элемент #root не найден')

createRoot(container).render(
  <StrictMode>
    <SWRConfig
      value={{
        // Сессия истекла или отозвана: выходим, а не показываем «Ошибка 401» на каждой странице.
        onError: (error) => {
          if (error instanceof ApiError && error.status === 401) signOut()
        },
        shouldRetryOnError: false,
      }}
    >
      <RouterProvider router={router} />
    </SWRConfig>
  </StrictMode>,
)
