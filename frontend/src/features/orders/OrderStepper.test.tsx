import { render, screen } from '@testing-library/react'

import { OrderStepper } from './OrderStepper'

describe('OrderStepper', () => {
  it('отмечает текущий шаг для оплаченного заказа', () => {
    render(<OrderStepper status="PAID" />)

    const steps = screen.getAllByRole('listitem')
    expect(steps).toHaveLength(3)
    expect(steps[1]).toHaveAttribute('aria-current', 'step')
    expect(steps[0]).not.toHaveAttribute('aria-current')
  })

  it('у собранного заказа нет текущего шага: путь завершён', () => {
    render(<OrderStepper status="ASSEMBLED" />)

    for (const step of screen.getAllByRole('listitem')) expect(step).not.toHaveAttribute('aria-current')
  })

  it('для отменённого заказа показывает сообщение вместо шагов', () => {
    render(<OrderStepper status="CANCELLED" />)

    expect(screen.getByText('Заказ отменён')).toBeInTheDocument()
    expect(screen.queryByRole('list')).not.toBeInTheDocument()
  })
})
