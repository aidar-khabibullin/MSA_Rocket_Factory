package model

import "errors"

var (
	// ErrPaymentNotFound — по заказу нет платежа, который можно вернуть.
	ErrPaymentNotFound = errors.New("payment not found")
	// ErrTransactionMismatch — транзакция в запросе на возврат не совпадает с платежом по заказу.
	ErrTransactionMismatch = errors.New("transaction does not match the order payment")
)
