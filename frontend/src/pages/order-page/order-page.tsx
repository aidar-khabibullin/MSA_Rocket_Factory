import { Spinner } from '@heroui/react/spinner'
import { useTranslation } from 'react-i18next'
import { useParams } from 'react-router-dom'

import { errorMessage } from '@/common/api/http'
import { AsyncGate } from '@/common/components/async-gate'
import { ErrorAlert } from '@/common/components/error-alert'
import { useOrder } from '@/modules/orders'

import { BackLink } from './back-link'
import { OrderDetails } from './order-details'

export const OrderPage = () => {
  const { t } = useTranslation('orders')
  const { orderId = '' } = useParams()
  const { data: order, error, isLoading } = useOrder(orderId)

  return (
    <AsyncGate
      data={order}
      isLoading={isLoading}
      loader={
        <div className="flex justify-center py-16">
          <Spinner aria-label={t('details.loading')} />
        </div>
      }
      fallback={
        <div className="mx-auto flex max-w-2xl flex-col gap-4">
          <BackLink />
          <ErrorAlert message={errorMessage(error, t('details.notFound'))} />
        </div>
      }
    >
      {(loaded) => <OrderDetails order={loaded} />}
    </AsyncGate>
  )
}
