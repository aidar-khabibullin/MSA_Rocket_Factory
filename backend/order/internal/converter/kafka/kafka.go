package kafka

import "github.com/aidar-khabibullin/MSA_Rocket_Factory/order/internal/model"

type ShipAssembledDecoder interface {
	Decode(data []byte) (*model.ShipAssembledEvent, error)
}
