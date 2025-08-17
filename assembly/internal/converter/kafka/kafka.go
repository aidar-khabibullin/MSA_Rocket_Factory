package kafka

import "github.com/aidar-khabibullin/MSA_Rocket_Factory/assembly/internal/model"

type AssemblyRecordedDecoder interface {
	Decode(data []byte) (model.AssemblyRecordedEvent, error)
}
