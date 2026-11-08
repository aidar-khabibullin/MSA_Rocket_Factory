package v1

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/aidar-khabibullin/MSA_Rocket_Factory/iam/internal/model"
	"github.com/aidar-khabibullin/MSA_Rocket_Factory/iam/internal/service"
	grpcMiddleware "github.com/aidar-khabibullin/MSA_Rocket_Factory/platform/pkg/middleware/grpc"
	iamV1 "github.com/aidar-khabibullin/MSA_Rocket_Factory/shared/pkg/proto/iam/v1"
)

// whoamiStub подменяет только Whoami: остальные методы AuthService в этих тестах не вызываются.
type whoamiStub struct {
	service.AuthService

	gotSession uuid.UUID
}

func (s *whoamiStub) Whoami(_ context.Context, sessionUUID uuid.UUID) (*model.Session, *model.User, error) {
	s.gotSession = sessionUUID
	return &model.Session{UUID: sessionUUID, ExpiresAt: time.Now().Add(time.Hour)}, &model.User{UUID: uuid.New(), Login: "demo"}, nil
}

func TestWhoami_SessionSource(t *testing.T) {
	sessionUUID := uuid.New()
	otherUUID := uuid.New()

	withMetadata := func(value string) context.Context {
		return metadata.NewIncomingContext(context.Background(), metadata.Pairs(grpcMiddleware.SessionUUIDMetadataKey, value))
	}

	tests := []struct {
		name     string
		ctx      context.Context
		req      *iamV1.WhoamiRequest
		wantUUID uuid.UUID
		wantCode codes.Code
	}{
		{"из поля запроса", context.Background(), &iamV1.WhoamiRequest{SessionUuid: sessionUUID.String()}, sessionUUID, codes.OK},
		{"из metadata, когда поле пустое", withMetadata(sessionUUID.String()), &iamV1.WhoamiRequest{}, sessionUUID, codes.OK},
		{"поле запроса важнее metadata", withMetadata(otherUUID.String()), &iamV1.WhoamiRequest{SessionUuid: sessionUUID.String()}, sessionUUID, codes.OK},
		{"нет ни поля, ни metadata", context.Background(), &iamV1.WhoamiRequest{}, uuid.Nil, codes.InvalidArgument},
		{"невалидный uuid в metadata", withMetadata("not-a-uuid"), &iamV1.WhoamiRequest{}, uuid.Nil, codes.InvalidArgument},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			stub := &whoamiStub{}
			res, err := NewAPI(stub).Whoami(tc.ctx, tc.req)

			require.Equal(t, tc.wantCode, status.Code(err))
			if tc.wantCode == codes.OK {
				require.NotNil(t, res.GetUser())
				require.Equal(t, tc.wantUUID, stub.gotSession)
			}
		})
	}
}
