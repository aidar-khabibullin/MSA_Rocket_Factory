import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'

import { ConfirmDialog } from './confirm-dialog'

function setup(props: { isOpen?: boolean; isPending?: boolean } = {}) {
  const onConfirm = vi.fn<() => void>()
  const onOpenChange = vi.fn<(isOpen: boolean) => void>()
  render(
    <ConfirmDialog
      isOpen={props.isOpen ?? true}
      isPending={props.isPending}
      onOpenChange={onOpenChange}
      onConfirm={onConfirm}
      title="Отменить заказ?"
      description="Вернуть не получится."
      confirmLabel="Да"
      cancelLabel="Нет"
    />,
  )
  return { onConfirm, onOpenChange }
}

describe('ConfirmDialog', () => {
  it('показывает заголовок и описание', () => {
    setup()
    expect(screen.getByRole('alertdialog', { name: 'Отменить заказ?' })).toBeInTheDocument()
    expect(screen.getByText('Вернуть не получится.')).toBeInTheDocument()
  })

  it('закрытый диалог ничего не рендерит', () => {
    setup({ isOpen: false })
    expect(screen.queryByRole('alertdialog')).not.toBeInTheDocument()
  })

  it('подтверждение вызывает onConfirm, отказ закрывает окно', async () => {
    const { onConfirm, onOpenChange } = setup()

    await userEvent.click(screen.getByRole('button', { name: 'Да' }))
    expect(onConfirm).toHaveBeenCalledTimes(1)

    await userEvent.click(screen.getByRole('button', { name: 'Нет' }))
    expect(onOpenChange).toHaveBeenCalledWith(false)
  })

  it('во время запроса кнопки заблокированы и Escape не закрывает окно', async () => {
    const { onConfirm, onOpenChange } = setup({ isPending: true })

    expect(screen.getByRole('button', { name: /Да/ })).toBeDisabled()
    expect(screen.getByRole('button', { name: 'Нет' })).toBeDisabled()

    await userEvent.keyboard('{Escape}')
    expect(onOpenChange).not.toHaveBeenCalled()
    expect(onConfirm).not.toHaveBeenCalled()
  })
})
