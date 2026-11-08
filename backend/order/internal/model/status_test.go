package model

import "testing"

func TestOrderStatus_CanTransitionTo(t *testing.T) {
	all := []OrderStatus{StatusPendingPayment, StatusPaid, StatusAssembled, StatusCancelled}

	allowed := map[[2]OrderStatus]bool{
		{StatusPendingPayment, StatusPaid}:      true,
		{StatusPendingPayment, StatusCancelled}: true,
		{StatusPaid, StatusAssembled}:           true,
	}

	for _, from := range all {
		for _, to := range all {
			want := allowed[[2]OrderStatus{from, to}]
			if got := from.CanTransitionTo(to); got != want {
				t.Errorf("%s -> %s: CanTransitionTo = %v, want %v", from, to, got, want)
			}
		}
	}
}
