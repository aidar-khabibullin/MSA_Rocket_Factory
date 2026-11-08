import { orderFromDto } from './order.adapter'

const dto = {
  order_uuid: 'o1',
  user_uuid: 'u1',
  part_uuids: ['p1', 'p1'],
  total_price: 200,
  transaction_uuid: 't1',
  payment_method: 'CARD',
  status: 'PAID',
}

describe('orderFromDto', () => {
  it('переводит поля бэкенда в camelCase-модель', () => {
    expect(orderFromDto(dto)).toEqual({
      uuid: 'o1',
      userUuid: 'u1',
      partUuids: ['p1', 'p1'],
      totalPrice: 200,
      transactionUuid: 't1',
      paymentMethod: 'CARD',
      status: 'PAID',
    })
  })

  it('нулевой UUID транзакции и UNKNOWN у неоплаченного заказа превращаются в отсутствие значения', () => {
    const order = orderFromDto({
      ...dto,
      transaction_uuid: '00000000-0000-0000-0000-000000000000',
      payment_method: 'UNKNOWN',
      status: 'PENDING_PAYMENT',
    })
    expect(order.transactionUuid).toBeUndefined()
    expect(order.paymentMethod).toBeUndefined()
  })

  it('неизвестный статус не ломает приложение', () => {
    expect(orderFromDto({ ...dto, status: 'REFUNDED' }).status).toBe('PENDING_PAYMENT')
  })
})
