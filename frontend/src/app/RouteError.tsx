import { Button } from '@heroui/react/button'
import { isRouteErrorResponse, useRouteError } from 'react-router-dom'

/** Граница ошибок роутера: сбой в странице или в загрузке чанка не должен давать белый экран. */
export function RouteError() {
  const error = useRouteError()
  const message = isRouteErrorResponse(error)
    ? error.status === 404
      ? 'Такой страницы нет.'
      : `Ошибка ${error.status}.`
    : 'Что-то пошло не так. Обновите страницу и попробуйте снова.'

  return (
    <div className="grid min-h-dvh place-items-center px-4">
      <div className="flex max-w-sm flex-col items-center gap-4 text-center" role="alert">
        <h1 className="text-2xl font-bold">Не удалось открыть страницу</h1>
        <p className="text-muted">{message}</p>
        <Button onPress={() => window.location.assign('/')}>На главную</Button>
      </div>
    </div>
  )
}
