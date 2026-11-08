package model

type OrderStatus string

const (
	StatusPendingPayment OrderStatus = "PENDING_PAYMENT"
	StatusPaid           OrderStatus = "PAID"
	StatusCancelled      OrderStatus = "CANCELLED"
	StatusAssembled      OrderStatus = "ASSEMBLED"
)

// transitions описывает допустимые переходы статусов заказа. Повтор того же статуса
// сюда не входит: он обрабатывается отдельно как идемпотентный no-op.
var transitions = map[OrderStatus][]OrderStatus{
	StatusPendingPayment: {StatusPaid, StatusCancelled},
	StatusPaid:           {StatusAssembled},
}

// CanTransitionTo сообщает, допустим ли переход из s в next.
// Устаревшее или повторное событие не должно откатывать заказ в более раннее состояние.
func (s OrderStatus) CanTransitionTo(next OrderStatus) bool {
	for _, allowed := range transitions[s] {
		if allowed == next {
			return true
		}
	}
	return false
}
