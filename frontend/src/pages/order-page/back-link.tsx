import ArrowLeft from '@gravity-ui/icons/ArrowLeft'
import { useTranslation } from 'react-i18next'
import { Link } from 'react-router-dom'

export const BackLink = () => {
  const { t } = useTranslation('orders')
  return (
    <Link to="/orders" className="inline-flex w-fit items-center gap-1.5 text-sm text-muted hover:text-foreground">
      <ArrowLeft className="size-4" aria-hidden="true" /> {t('details.back')}
    </Link>
  )
}
