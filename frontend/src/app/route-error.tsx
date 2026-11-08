import { Button } from '@heroui/react/button'
import { useTranslation } from 'react-i18next'
import { isRouteErrorResponse, useRouteError } from 'react-router-dom'

/** Граница ошибок роутера: сбой в странице или в загрузке чанка не должен давать белый экран. */
export const RouteError = ({ inline = false }: { inline?: boolean }) => {
  const { t } = useTranslation()
  const error = useRouteError()
  const message = isRouteErrorResponse(error)
    ? error.status === 404
      ? t('errors.routeNotFound')
      : t('errors.status', { status: error.status })
    : t('errors.routeGeneric')

  return (
    <div className={`grid place-items-center px-4 ${inline ? 'py-16' : 'min-h-dvh'}`}>
      <div className="flex max-w-sm flex-col items-center gap-4 text-center" role="alert">
        <h1 className="text-2xl font-bold">{t('errors.routeTitle')}</h1>
        <p className="text-muted">{message}</p>
        <div className="flex gap-2">
          {inline ? (
            <Button variant="tertiary" onPress={() => window.location.reload()}>
              {t('errors.retry')}
            </Button>
          ) : null}
          <Button onPress={() => window.location.assign('/')}>{t('errors.home')}</Button>
        </div>
      </div>
    </div>
  )
}
