import { Button } from '@heroui/react/button'
import { Spinner } from '@heroui/react/spinner'
import { useTranslation } from 'react-i18next'

import { errorMessage } from '@/common/api/http'
import { ConfirmDialog } from '@/common/components/confirm-dialog'
import { ErrorAlert } from '@/common/components/error-alert'
import { formatPrice } from '@/common/lib/format'
import { PaymentMethodPicker, type Order } from '@/modules/orders'

import { usePaymentPanel } from './use-payment-panel'

interface PaymentPanelProps {
  order: Order
}

export const PaymentPanel = ({ order }: PaymentPanelProps) => {
  const { t } = useTranslation('orders')
  const panel = usePaymentPanel(order.uuid)

  return (
    <section className="flex flex-col gap-4 rounded-2xl border border-accent/40 bg-surface p-6" aria-label={t('pay.region')}>
      <PaymentMethodPicker value={panel.method} onChange={panel.setMethod} />

      <ErrorAlert message={panel.error ? errorMessage(panel.error, t('pay.failed')) : null} />

      <div className="flex flex-wrap gap-2">
        <Button className="flex-1" isDisabled={panel.isBusy} onPress={() => void panel.payOrder()}>
          {panel.isPaying ? <Spinner size="sm" color="current" /> : null}
          {t('pay.pay')}
          {order.totalPrice === undefined ? '' : ` ${formatPrice(order.totalPrice)}`}
        </Button>
        <Button variant="danger-soft" isDisabled={panel.isBusy} onPress={() => panel.setCancelDialogOpen(true)}>
          {t('pay.cancel')}
        </Button>
      </div>

      <ConfirmDialog
        isOpen={panel.isCancelDialogOpen}
        onOpenChange={panel.setCancelDialogOpen}
        title={t('pay.confirmTitle')}
        description={t('pay.confirmText')}
        confirmLabel={t('pay.confirmAction')}
        cancelLabel={t('pay.keep')}
        onConfirm={() => void panel.confirmCancel()}
        isPending={panel.isCancelling}
      />
    </section>
  )
}
