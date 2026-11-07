package inmemory_test

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/google/uuid"

	"github.com/aidar-khabibullin/MSA_Rocket_Factory/inventory/internal/model"
)

var (
	partA = uuid.MustParse("d973e963-b7e6-4323-8f4e-4bfd5ab8e834") // остаток 10
	partB = uuid.MustParse("d973e963-b7e6-4323-8f4e-4bfd5ab8e835") // остаток 20
)

func (s *InMemoryRepositorySuite) stock(id uuid.UUID) int64 {
	part, err := s.repository.GetPart(context.Background(), id)
	s.Require().NoError(err)
	return part.StockQuantity
}

func (s *InMemoryRepositorySuite) TestReserve_DecrementsStockPerPartCount() {
	err := s.repository.Reserve(context.Background(), uuid.New(), []uuid.UUID{partA, partA, partB})

	s.Require().NoError(err)
	s.Equal(int64(8), s.stock(partA), "повторяющийся uuid означает несколько штук")
	s.Equal(int64(19), s.stock(partB))
}

func (s *InMemoryRepositorySuite) TestReserve_IsIdempotentPerOrder() {
	order := uuid.New()

	s.Require().NoError(s.repository.Reserve(context.Background(), order, []uuid.UUID{partA}))
	s.Require().NoError(s.repository.Reserve(context.Background(), order, []uuid.UUID{partA}))

	s.Equal(int64(9), s.stock(partA), "повторный резерв того же заказа не должен списывать остаток второй раз")
}

func (s *InMemoryRepositorySuite) TestReserve_InsufficientStockReservesNothing() {
	// partB доступен, но 11 штук partA больше остатка: не должно зарезервироваться ничего.
	parts := []uuid.UUID{partB}
	for range 11 {
		parts = append(parts, partA)
	}

	err := s.repository.Reserve(context.Background(), uuid.New(), parts)

	s.ErrorIs(err, model.ErrInsufficientStock)
	s.Equal(int64(10), s.stock(partA))
	s.Equal(int64(20), s.stock(partB), "резерв атомарен: деталь, которой хватало, тоже не списывается")
}

func (s *InMemoryRepositorySuite) TestReserve_UnknownPart() {
	err := s.repository.Reserve(context.Background(), uuid.New(), []uuid.UUID{partA, uuid.New()})

	s.ErrorIs(err, model.ErrPartNotFound)
	s.Equal(int64(10), s.stock(partA))
}

func (s *InMemoryRepositorySuite) TestReserve_EmptyOrderIsRejected() {
	s.ErrorIs(s.repository.Reserve(context.Background(), uuid.New(), nil), model.ErrEmptyReservation)
}

func (s *InMemoryRepositorySuite) TestRelease_ReturnsStockAndIsIdempotent() {
	order := uuid.New()
	s.Require().NoError(s.repository.Reserve(context.Background(), order, []uuid.UUID{partA, partA}))

	s.Require().NoError(s.repository.Release(context.Background(), order))
	s.Require().NoError(s.repository.Release(context.Background(), order))

	s.Equal(int64(10), s.stock(partA), "остаток вернулся ровно один раз")
}

func (s *InMemoryRepositorySuite) TestRelease_UnknownOrderIsNoop() {
	s.NoError(s.repository.Release(context.Background(), uuid.New()))
	s.Equal(int64(10), s.stock(partA))
}

func (s *InMemoryRepositorySuite) TestReserve_AfterReleaseOrderCanReserveAgain() {
	order := uuid.New()
	s.Require().NoError(s.repository.Reserve(context.Background(), order, []uuid.UUID{partA}))
	s.Require().NoError(s.repository.Release(context.Background(), order))

	s.Require().NoError(s.repository.Reserve(context.Background(), order, []uuid.UUID{partA}))

	s.Equal(int64(9), s.stock(partA))
}

// 30 заказов борются за 10 штук: ровно 10 успешных, остаток никогда не уходит в минус.
func (s *InMemoryRepositorySuite) TestReserve_ConcurrentOrdersNeverOversell() {
	var ok, rejected atomic.Int64
	var wg sync.WaitGroup

	for range 30 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := s.repository.Reserve(context.Background(), uuid.New(), []uuid.UUID{partA})
			switch err {
			case nil:
				ok.Add(1)
			default:
				s.ErrorIs(err, model.ErrInsufficientStock)
				rejected.Add(1)
			}
		}()
	}
	wg.Wait()

	s.Equal(int64(10), ok.Load())
	s.Equal(int64(20), rejected.Load())
	s.Equal(int64(0), s.stock(partA))
}
