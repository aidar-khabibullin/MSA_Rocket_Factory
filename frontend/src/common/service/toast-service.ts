import { toast } from '@heroui/react/toast'

/** Контракт уведомлений: код приложения зависит от него, а не от конкретной библиотеки тостов. */
export interface Toaster {
  showSuccess: (message: string) => void
  showError: (message: string) => void
  showWarning: (message: string) => void
  showInfo: (message: string) => void
}

export const toastService: Toaster = {
  showSuccess: (message) => void toast.success(message),
  showError: (message) => void toast.danger(message),
  showWarning: (message) => void toast.warning(message),
  showInfo: (message) => void toast.info(message),
}
