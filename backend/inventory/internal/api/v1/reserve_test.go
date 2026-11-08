package v1

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/aidar-khabibullin/MSA_Rocket_Factory/inventory/internal/model"
	"github.com/aidar-khabibullin/MSA_Rocket_Factory/platform/pkg/logger"
	inventoryV1 "github.com/aidar-khabibullin/MSA_Rocket_Factory/shared/pkg/proto/inventory/v1"
)

func TestReserveParts_Success(t *testing.T) {
	_ = logger.InitSimple("info", false)
	svc := new(MockInventoryService)
	order, a, b := uuid.New(), uuid.New(), uuid.New()
	svc.On("ReserveParts", mock.Anything, order, []uuid.UUID{a, b}).Return(nil)

	_, err := NewAPI(svc).ReserveParts(context.Background(), &inventoryV1.ReservePartsRequest{
		OrderUuid: order.String(), PartUuids: []string{a.String(), b.String()},
	})

	require.NoError(t, err)
	svc.AssertExpectations(t)
}

func TestReserveParts_ErrorMapping(t *testing.T) {
	_ = logger.InitSimple("info", false)
	valid := uuid.New().String()
	tests := []struct {
		name     string
		order    string
		parts    []string
		svcErr   error
		wantCode codes.Code
	}{
		{"невалидный order_uuid", "x", []string{valid}, nil, codes.InvalidArgument},
		{"невалидный part_uuid", valid, []string{"x"}, nil, codes.InvalidArgument},
		{"пустой резерв", valid, nil, model.ErrEmptyReservation, codes.InvalidArgument},
		{"деталь не найдена", valid, []string{valid}, model.ErrPartNotFound, codes.NotFound},
		{"нет остатка", valid, []string{valid}, model.ErrInsufficientStock, codes.FailedPrecondition},
		{"прочая ошибка", valid, []string{valid}, errors.New("boom"), codes.Internal},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := new(MockInventoryService)
			if tt.svcErr != nil {
				svc.On("ReserveParts", mock.Anything, mock.Anything, mock.Anything).Return(tt.svcErr)
			}

			_, err := NewAPI(svc).ReserveParts(context.Background(), &inventoryV1.ReservePartsRequest{OrderUuid: tt.order, PartUuids: tt.parts})

			require.Equal(t, tt.wantCode, status.Code(err))
		})
	}
}

func TestReleaseParts(t *testing.T) {
	_ = logger.InitSimple("info", false)
	order := uuid.New()

	svc := new(MockInventoryService)
	svc.On("ReleaseParts", mock.Anything, order).Return(nil).Once()
	_, err := NewAPI(svc).ReleaseParts(context.Background(), &inventoryV1.ReleasePartsRequest{OrderUuid: order.String()})
	require.NoError(t, err)

	svc.On("ReleaseParts", mock.Anything, order).Return(errors.New("boom")).Once()
	_, err = NewAPI(svc).ReleaseParts(context.Background(), &inventoryV1.ReleasePartsRequest{OrderUuid: order.String()})
	require.Equal(t, codes.Internal, status.Code(err))

	_, err = NewAPI(svc).ReleaseParts(context.Background(), &inventoryV1.ReleasePartsRequest{OrderUuid: "x"})
	require.Equal(t, codes.InvalidArgument, status.Code(err))
}
