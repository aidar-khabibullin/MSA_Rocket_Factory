package mongo

import (
	"context"

	"github.com/aidar-khabibullin/MSA_Rocket_Factory/inventory/internal/model"
)

func (r *repository) ListParts(ctx context.Context, filter *model.Filter) (*[]model.Part, error) {
	//TODO implement me
	return nil, nil
}
