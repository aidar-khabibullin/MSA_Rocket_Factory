package kafka

import "github.com/aidar-khabibullin/MSA_Rocket_Factory/notification/internal/model"

type OrderPaidDecoder interface {
	Decode(data []byte) (*model.OrderPaidEvent, error)
}

type ShipAssembledDecoder interface {
	Decode(data []byte) (*model.ShipAssembledEvent, error)
}
