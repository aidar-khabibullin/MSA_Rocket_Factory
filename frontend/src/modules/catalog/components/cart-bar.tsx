import ShoppingCart from '@gravity-ui/icons/ShoppingCart'
import { Button } from '@heroui/react/button'
import { Spinner } from '@heroui/react/spinner'
import { useTranslation } from 'react-i18next'

import { formatPrice } from '@/common/lib/format'

interface CartBarProps {
  count: number
  total: number
  isPending: boolean
  onClear: () => void
  onCheckout: () => void
}

export function CartBar({ count, total, isPending, onClear, onCheckout }: CartBarProps) {
  const { t } = useTranslation('catalog')
  return (
    <div className="fixed inset-x-0 bottom-4 z-30 px-4">
      <div className="rise mx-auto flex max-w-3xl items-center justify-between gap-4 rounded-2xl border border-accent/40 bg-overlay/90 p-3 pl-5 shadow-2xl backdrop-blur-md">
        <div className="flex items-center gap-3">
          <span className="grid size-10 place-items-center rounded-xl bg-accent/15 text-accent" aria-hidden="true">
            <ShoppingCart />
          </span>
          <div>
            <p className="text-sm text-muted">
              {t('cart.count')} <b className="text-foreground">{count}</b>
            </p>
            <p className="text-lg font-bold leading-tight tabular-nums">{formatPrice(total)}</p>
          </div>
        </div>
        <div className="flex gap-2">
          <Button variant="ghost" onPress={onClear} isDisabled={isPending}>
            {t('cart.clear')}
          </Button>
          <Button onPress={onCheckout} isDisabled={isPending}>
            {isPending ? <Spinner size="sm" color="current" /> : null}
            {t('cart.checkout')}
          </Button>
        </div>
      </div>
    </div>
  )
}
