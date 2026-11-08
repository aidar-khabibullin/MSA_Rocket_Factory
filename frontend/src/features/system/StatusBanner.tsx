import { Button } from '@heroui/react/button'

import { useServiceStatus } from './useServiceStatus'

/** Для пользователя виден только факт проблемы, без названий внутренних сервисов. */
export function StatusBanner() {
  const { isDegraded, isChecking, refresh } = useServiceStatus()
  if (!isDegraded) return null

  return (
    <output className="block border-b border-warning/40 bg-warning/10">
      <div className="mx-auto flex max-w-6xl flex-wrap items-center justify-between gap-3 px-4 py-3 text-sm sm:px-6">
        <p>Часть функций временно недоступна. Попробуйте повторить действие чуть позже.</p>
        <Button size="sm" variant="outline" onPress={refresh} isDisabled={isChecking}>
          Проверить снова
        </Button>
      </div>
    </output>
  )
}
