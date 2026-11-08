import { isAwaitingAssembly, isRealUuid, PROGRESS_STEPS, reachedStep } from './status'

describe('reachedStep', () => {
  it('сопоставляет статус с последним достигнутым шагом', () => {
    expect(reachedStep('PENDING_PAYMENT')).toBe(0)
    expect(reachedStep('PAID')).toBe(1)
    expect(reachedStep('ASSEMBLED')).toBe(PROGRESS_STEPS.length - 1)
  })

  it('для отменённого заказа шагов нет', () => {
    expect(reachedStep('CANCELLED')).toBeNull()
  })
})

describe('isAwaitingAssembly', () => {
  it('опрос нужен только пока заказ оплачен и ещё не собран', () => {
    expect(isAwaitingAssembly('PAID')).toBe(true)
    expect(isAwaitingAssembly('ASSEMBLED')).toBe(false)
    expect(isAwaitingAssembly('PENDING_PAYMENT')).toBe(false)
    expect(isAwaitingAssembly('CANCELLED')).toBe(false)
    expect(isAwaitingAssembly(undefined)).toBe(false)
  })
})

describe('isRealUuid', () => {
  it('отсекает пустое значение и нулевой UUID бэкенда', () => {
    expect(isRealUuid(undefined)).toBe(false)
    expect(isRealUuid('')).toBe(false)
    expect(isRealUuid('00000000-0000-0000-0000-000000000000')).toBe(false)
    expect(isRealUuid('d973e963-b7e6-4323-8f4e-4bfd5ab8e834')).toBe(true)
  })
})
