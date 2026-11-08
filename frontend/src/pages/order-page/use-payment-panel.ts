import { useState } from 'react'
import { useTranslation } from 'react-i18next'

import { toastService } from '@/common/service/toast-service'
import { useCancelOrder, usePayOrder, type PaymentMethod } from '@/modules/orders'

/** Состояние и действия панели оплаты: выбор способа, оплата, отмена с подтверждением. */
export function usePaymentPanel(orderUuid: string) {
  const { t } = useTranslation('orders')
  const [method, setMethod] = useState<PaymentMethod>('CARD')
  const [isCancelDialogOpen, setCancelDialogOpen] = useState(false)
  const pay = usePayOrder(orderUuid)
  const cancel = useCancelOrder(orderUuid)

  async function run(action: () => Promise<unknown>, successMessage: string): Promise<boolean> {
    try {
      await action()
      toastService.showSuccess(successMessage)
      return true
    } catch {
      return false // ошибка сохранена в состоянии мутации и показывается панелью
    }
  }

  return {
    method,
    setMethod,
    isPaying: pay.isPending,
    isCancelling: cancel.isPending,
    isBusy: pay.isPending || cancel.isPending,
    error: pay.error ?? cancel.error,
    isCancelDialogOpen,
    setCancelDialogOpen,
    payOrder: () => run(() => pay.mutateAsync(method), t('pay.paid')),
    // Окно закрываем в любом случае: при ошибке она видна в панели, а не за затемнением.
    confirmCancel: async () => {
      await run(() => cancel.mutateAsync(), t('pay.cancelled'))
      setCancelDialogOpen(false)
    },
  }
}
