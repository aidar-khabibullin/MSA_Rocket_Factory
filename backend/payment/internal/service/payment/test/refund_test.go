package payment_test

import (
	"sync"

	"github.com/google/uuid"

	"github.com/aidar-khabibullin/MSA_Rocket_Factory/payment/internal/model"
)

func order() model.Order {
	return model.Order{OrderUuid: uuid.New(), UserUuid: uuid.New(), PaymentMethod: "CARD"}
}

func (s *ServiceSuite) TestPay_IsIdempotentPerOrder() {
	o := order()

	first, err := s.service.Pay(s.T().Context(), o)
	s.Require().NoError(err)
	second, err := s.service.Pay(s.T().Context(), o)
	s.Require().NoError(err)

	s.Equal(first, second, "повторная оплата того же заказа должна вернуть ту же транзакцию, а не списать деньги дважды")
}

func (s *ServiceSuite) TestPay_DifferentOrdersGetDifferentTransactions() {
	a, err := s.service.Pay(s.T().Context(), order())
	s.Require().NoError(err)
	b, err := s.service.Pay(s.T().Context(), order())
	s.Require().NoError(err)

	s.NotEqual(a, b)
}

func (s *ServiceSuite) TestPay_ConcurrentCallsForSameOrderProduceOneTransaction() {
	o := order()
	results := make([]uuid.UUID, 50)

	var wg sync.WaitGroup
	for i := range results {
		wg.Add(1)
		go func() {
			defer wg.Done()
			tx, err := s.service.Pay(s.T().Context(), o)
			s.NoError(err)
			results[i] = tx
		}()
	}
	wg.Wait()

	for _, tx := range results {
		s.Equal(results[0], tx)
	}
}

func (s *ServiceSuite) TestRefund_ReturnsPaymentAndIsIdempotent() {
	o := order()
	tx, err := s.service.Pay(s.T().Context(), o)
	s.Require().NoError(err)

	first, err := s.service.Refund(s.T().Context(), o.OrderUuid, tx)
	s.Require().NoError(err)
	second, err := s.service.Refund(s.T().Context(), o.OrderUuid, tx)
	s.Require().NoError(err)

	s.NotEqual(uuid.Nil, first)
	s.Equal(first, second, "повторный возврат не должен создавать второй возврат")
}

func (s *ServiceSuite) TestRefund_UnknownOrder() {
	_, err := s.service.Refund(s.T().Context(), uuid.New(), uuid.New())

	s.ErrorIs(err, model.ErrPaymentNotFound)
}

func (s *ServiceSuite) TestRefund_WrongTransaction() {
	o := order()
	_, err := s.service.Pay(s.T().Context(), o)
	s.Require().NoError(err)

	_, err = s.service.Refund(s.T().Context(), o.OrderUuid, uuid.New())

	s.ErrorIs(err, model.ErrTransactionMismatch)
}

func (s *ServiceSuite) TestPay_AfterRefundCreatesNewTransaction() {
	o := order()
	first, _ := s.service.Pay(s.T().Context(), o)
	_, err := s.service.Refund(s.T().Context(), o.OrderUuid, first)
	s.Require().NoError(err)

	second, err := s.service.Pay(s.T().Context(), o)

	s.Require().NoError(err)
	s.NotEqual(first, second, "после возврата заказ можно оплатить снова, это новая транзакция")
}
