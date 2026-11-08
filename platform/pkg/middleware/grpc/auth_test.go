package grpc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func callUnary(t *testing.T, fullMethod string) (called bool, err error) {
	t.Helper()
	// iamClient не нужен: без session-uuid в metadata аутентификация отклоняется до обращения к IAM.
	interceptor := NewAuthInterceptor(nil).Unary()

	_, err = interceptor(context.Background(), nil, &grpc.UnaryServerInfo{FullMethod: fullMethod},
		func(context.Context, any) (any, error) {
			called = true
			return nil, nil
		})
	return called, err
}

func TestAuthInterceptor_HealthCheckDoesNotRequireSession(t *testing.T) {
	for _, method := range []string{"/grpc.health.v1.Health/Check", "/grpc.health.v1.Health/Watch"} {
		called, err := callUnary(t, method)

		require.NoError(t, err, method)
		require.True(t, called, method)
	}
}

func TestAuthInterceptor_BusinessMethodRequiresSession(t *testing.T) {
	called, err := callUnary(t, "/inventory.v1.InventoryService/ListParts")

	require.False(t, called)
	require.Equal(t, codes.Unauthenticated, status.Code(err))
}

func TestAuthInterceptor_PrefixDoesNotMatchLookalikeServices(t *testing.T) {
	called, err := callUnary(t, "/grpc.health.v1.HealthExtra/Check")

	require.False(t, called)
	require.Equal(t, codes.Unauthenticated, status.Code(err))
}
