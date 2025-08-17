package service

import (
	"context"

	"github.com/aidar-khabibullin/MSA_Rocket_Factory/assembly/internal/model"
)

type AssemblyService interface {
	Assemble(ctx context.Context, event model.OrderPaidEvent) error
}

type ConsumerService interface {
	RunConsumer(ctx context.Context) error
}

type ProducerService interface {
	ProduceAssembly(ctx context.Context, event model.ShipAssembledEvent) error
}
