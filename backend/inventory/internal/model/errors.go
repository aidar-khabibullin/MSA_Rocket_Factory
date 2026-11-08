package model

import "errors"

var (
	ErrPartNotFound    = errors.New("part not found")
	ErrConvertFromRepo = errors.New("can't parse to model")
	// ErrInsufficientStock — на складе нет нужного количества хотя бы одной детали.
	ErrInsufficientStock = errors.New("insufficient stock")
	// ErrEmptyReservation — в резерве нет ни одной детали.
	ErrEmptyReservation = errors.New("reservation must contain at least one part")
)
