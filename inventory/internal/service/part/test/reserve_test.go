package part_test

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/aidar-khabibullin/MSA_Rocket_Factory/inventory/internal/model"
)

func (s *ServiceSuite) TestReserveParts_PassesOrderAndPartsToRepository() {
	order, a, b := uuid.New(), uuid.New(), uuid.New()
	s.inventoryRepo.On("Reserve", mock.Anything, order, []uuid.UUID{a, b, a}).Return(nil).Once()

	s.Require().NoError(s.service.ReserveParts(context.Background(), order, []uuid.UUID{a, b, a}))
}

func (s *ServiceSuite) TestReserveParts_KeepsDomainErrorForMapping() {
	for _, repoErr := range []error{model.ErrInsufficientStock, model.ErrPartNotFound, model.ErrEmptyReservation} {
		s.inventoryRepo.On("Reserve", mock.Anything, mock.Anything, mock.Anything).Return(repoErr).Once()

		err := s.service.ReserveParts(context.Background(), uuid.New(), []uuid.UUID{uuid.New()})

		s.ErrorIs(err, repoErr, "API-слой должен суметь отличить причину отказа")
	}
}

func (s *ServiceSuite) TestReleaseParts() {
	order := uuid.New()
	s.inventoryRepo.On("Release", mock.Anything, order).Return(nil).Once()
	s.Require().NoError(s.service.ReleaseParts(context.Background(), order))

	boom := errors.New("db down")
	s.inventoryRepo.On("Release", mock.Anything, order).Return(boom).Once()
	s.ErrorIs(s.service.ReleaseParts(context.Background(), order), boom)
}
