package v1

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/aidar-khabibullin/MSA_Rocket_Factory/iam/internal/model"
	iamV1 "github.com/aidar-khabibullin/MSA_Rocket_Factory/shared/pkg/proto/iam/v1"
)

type stubUserService struct {
	err error
}

func (s stubUserService) Register(context.Context, *model.UserRegistrationInfo) (*model.User, error) {
	if s.err != nil {
		return nil, s.err
	}
	return &model.User{UUID: uuid.New()}, nil
}

func (stubUserService) GetUser(context.Context, uuid.UUID) (*model.User, error) { return nil, nil }

func registerRequest(password string) *iamV1.RegisterRequest {
	return &iamV1.RegisterRequest{Info: &iamV1.UserRegistrationInfo{
		UserInfo: &iamV1.UserInfo{Login: "valid_login", Email: "a@b.io"},
		Password: password,
	}}
}

func TestRegister_ErrorMapping(t *testing.T) {
	weak := model.ValidatePassword("weak")
	require.Error(t, weak)

	tests := []struct {
		name     string
		err      error
		wantCode codes.Code
		wantMsg  string
	}{
		{"слабый пароль -> InvalidArgument с причиной", fmt.Errorf("ошибка валидации пароля: %w", weak), codes.InvalidArgument, "минимум 8 символов"},
		{"некорректный логин -> InvalidArgument", model.ValidateLogin("a"), codes.InvalidArgument, "логин"},
		{"некорректный email -> InvalidArgument", model.ValidateEmail("x"), codes.InvalidArgument, "invalid email format"},
		{"пользователь существует -> AlreadyExists", model.ErrUserAlreadyExists, codes.AlreadyExists, ""},
		{"прочее -> Internal", fmt.Errorf("db down"), codes.Internal, "internal server error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewAPI(stubUserService{err: tt.err}).Register(context.Background(), registerRequest("x"))
			st, ok := status.FromError(err)
			require.True(t, ok)
			require.Equal(t, tt.wantCode, st.Code())
			require.Contains(t, st.Message(), tt.wantMsg)
		})
	}
}

func TestRegister_Success(t *testing.T) {
	res, err := NewAPI(stubUserService{}).Register(context.Background(), registerRequest("Secret123!"))
	require.NoError(t, err)
	require.NotEmpty(t, res.GetUserUuid())
}
