import { Button } from '@heroui/react/button'
import { useTranslation } from 'react-i18next'

import { useServiceStatus } from '../hooks/use-service-status'

/** Для пользователя виден только факт проблемы, без названий внутренних сервисов. */
export function StatusBanner() {
  const { t } = useTranslation('system')
  const { isDegraded, isChecking, refresh } = useServiceStatus()
  if (!isDegraded) return null

  return (
    <output className="block border-b border-warning/40 bg-warning/10">
      <div className="mx-auto flex max-w-6xl flex-wrap items-center justify-between gap-3 px-4 py-3 text-sm sm:px-6">
        <p>{t('banner.text')}</p>
        <Button size="sm" variant="outline" onPress={refresh} isDisabled={isChecking}>
          {t('banner.retry')}
        </Button>
      </div>
    </output>
  )
}
