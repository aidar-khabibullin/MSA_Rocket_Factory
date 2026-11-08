import './locales'

// Публичный API модуля: снаружи импортируем только отсюда, а не из внутренних папок.
export { OrderComposition } from './components/order-composition'
export { OrderStepper } from './components/order-stepper'
export { PaymentMethodPicker } from './components/payment-method-picker'
export { useCancelOrder, useCreateOrder, useOrder, useOrders, usePayOrder } from './hooks/use-orders'
export { groupPartUuids } from './model/group-part-uuids'
export type { Order, PaymentMethod } from './model/order'
export { isAwaitingAssembly, STATUS_TONE } from './model/status'
export { rememberOrder, useOrderHistory } from './store/order-history.store'
