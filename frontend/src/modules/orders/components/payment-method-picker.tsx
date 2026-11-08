import { useTranslation } from 'react-i18next'

import { PAYMENT_METHODS, type PaymentMethod } from '../model/order'

interface PaymentMethodPickerProps {
  value: PaymentMethod
  onChange: (method: PaymentMethod) => void
}

/** Нативные radio внутри fieldset: клавиатура и скринридеры работают без дополнительного кода. */
export const PaymentMethodPicker = ({ value, onChange }: PaymentMethodPickerProps) => {
  const { t } = useTranslation('orders')
  return (
    <fieldset className="flex flex-col gap-3">
      <legend className="mb-1 font-bold">{t('payment.title')}</legend>
      <div className="grid gap-2 sm:grid-cols-2">
        {PAYMENT_METHODS.map((method) => (
          <label
            key={method}
            className="flex min-h-12 cursor-pointer items-center gap-3 rounded-xl border border-border px-4 transition-colors hover:border-accent/50 has-[:checked]:border-accent has-[:checked]:bg-accent/10 has-[:focus-visible]:outline-2 has-[:focus-visible]:outline-accent"
          >
            <input
              type="radio"
              name="payment-method"
              value={method}
              checked={value === method}
              onChange={() => onChange(method)}
              className="size-4 accent-[var(--accent)]"
            />
            <span className="text-sm font-medium">{t(`payment.${method}`)}</span>
          </label>
        ))}
      </div>
    </fieldset>
  )
}
