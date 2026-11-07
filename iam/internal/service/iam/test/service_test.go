package iam_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/aidar-khabibullin/MSA_Rocket_Factory/iam/internal/model"
	"github.com/aidar-khabibullin/MSA_Rocket_Factory/iam/internal/repository/mocks"
	"github.com/aidar-khabibullin/MSA_Rocket_Factory/iam/internal/service/iam"
	"github.com/aidar-khabibullin/MSA_Rocket_Factory/platform/pkg/logger"
)

type jwtConfig struct{}

func (jwtConfig) AccessTokenSecret() string      { return "access" }
func (jwtConfig) RefreshTokenSecret() string     { return "refresh" }
func (jwtConfig) AccessTokenTTL() time.Duration  { return time.Minute }
func (jwtConfig) RefreshTokenTTL() time.Duration { return time.Hour }

func TestFacadeDelegatesToUserAndAuthServices(t *testing.T) {
	logger.SetNopLogger()
	ctx := context.Background()
	repo := mocks.NewIAMRepository(t)
	svc := iam.NewService(repo, jwtConfig{})

	user := &model.User{UUID: uuid.New(), Login: "rocket"}
	session := &model.Session{UUID: uuid.New(), UserUUID: user.UUID, ExpiresAt: time.Now().Add(time.Hour)}

	// GetUser -> userService
	repo.On("GetUserByUUID", mock.Anything, user.UUID).Return(user, nil)
	got, err := svc.GetUser(ctx, user.UUID)
	require.NoError(t, err)
	require.Equal(t, user, got)

	// Whoami -> authService (сессия берётся из кеша)
	repo.On("GetSessionFromCache", mock.Anything, session.UUID).Return(session, nil)
	gotSession, gotUser, err := svc.Whoami(ctx, session.UUID)
	require.NoError(t, err)
	require.Equal(t, session, gotSession)
	require.Equal(t, user, gotUser)

	// Logout -> authService
	repo.On("DeleteSession", mock.Anything, session.UUID).Return(nil)
	repo.On("Delete", mock.Anything, session.UUID).Return(nil)
	require.NoError(t, svc.Logout(ctx, session.UUID))
}

func TestFacadeValidationErrorsComeFromServices(t *testing.T) {
	logger.SetNopLogger()
	ctx := context.Background()
	svc := iam.NewService(mocks.NewIAMRepository(t), jwtConfig{})

	_, err := svc.Login(ctx, "", "password")
	require.Error(t, err)

	_, err = svc.JWTLogin(ctx, "", "password")
	require.Error(t, err)

	_, err = svc.GetAccessToken(ctx, "not-a-token")
	require.Error(t, err)

	_, err = svc.GetRefreshToken(ctx, "not-a-token")
	require.Error(t, err)

	_, err = svc.Register(ctx, &model.UserRegistrationInfo{})
	require.Error(t, err)
}
