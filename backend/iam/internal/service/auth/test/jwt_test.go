package auth_test

import (
	"context"
	"errors"
	"time"

	"github.com/stretchr/testify/mock"

	"github.com/aidar-khabibullin/MSA_Rocket_Factory/iam/internal/model"
	"github.com/aidar-khabibullin/MSA_Rocket_Factory/iam/internal/repository/mocks"
	"github.com/aidar-khabibullin/MSA_Rocket_Factory/iam/internal/service/auth"
)

// expectNotRevoked настраивает blacklist-проверку: токен не найден в кеше
func (s *AuthServiceSuite) expectNotRevoked() {
	s.repo.On("GetSessionByUUID", mock.Anything, mock.AnythingOfType("uuid.UUID")).
		Return(nil, model.ErrSessionNotFound)
}

func (s *AuthServiceSuite) jwtLogin() *model.TokenPair {
	s.repo.On("GetUserByLogin", mock.Anything, testLogin).Return(s.newUser(), nil)

	pair, err := s.service.JWTLogin(context.Background(), testLogin, testPassword)
	s.Require().NoError(err)
	s.Require().NotNil(pair)

	return pair
}

func (s *AuthServiceSuite) TestJWTLogin_Validation() {
	_, err := s.service.JWTLogin(context.Background(), "", testPassword)
	s.Require().ErrorIs(err, model.ErrEmptyLogin)

	_, err = s.service.JWTLogin(context.Background(), testLogin, "")
	s.Require().ErrorIs(err, model.ErrEmptyPassword)
}

func (s *AuthServiceSuite) TestJWTLogin_UserNotFound() {
	s.repo.On("GetUserByLogin", mock.Anything, testLogin).Return(nil, model.ErrUserNotFound)

	_, err := s.service.JWTLogin(context.Background(), testLogin, testPassword)

	s.Require().ErrorIs(err, model.ErrInvalidCredentials)
}

func (s *AuthServiceSuite) TestJWTLogin_WrongPassword() {
	s.repo.On("GetUserByLogin", mock.Anything, testLogin).Return(s.newUser(), nil)

	_, err := s.service.JWTLogin(context.Background(), testLogin, "Wrong!Passw0rd")

	s.Require().ErrorIs(err, model.ErrInvalidCredentials)
}

func (s *AuthServiceSuite) TestJWTLogin_CorruptedHash() {
	u := s.newUser()
	u.PasswordHash = "$argon2id$broken"
	s.repo.On("GetUserByLogin", mock.Anything, testLogin).Return(u, nil)

	_, err := s.service.JWTLogin(context.Background(), testLogin, testPassword)

	s.Require().ErrorIs(err, model.ErrPasswordVerification)
}

func (s *AuthServiceSuite) TestJWTLogin_Success() {
	pair := s.jwtLogin()

	s.NotEmpty(pair.AccessToken)
	s.NotEmpty(pair.RefreshToken)
	s.NotEqual(pair.AccessToken, pair.RefreshToken)
	s.True(pair.AccessTokenExpiresAt.Before(pair.RefreshTokenExpiresAt))
}

func (s *AuthServiceSuite) TestGetAccessToken_EmptyToken() {
	_, err := s.service.GetAccessToken(context.Background(), "")

	s.Require().ErrorIs(err, auth.ErrInvalidToken)
}

func (s *AuthServiceSuite) TestGetAccessToken_GarbageToken() {
	_, err := s.service.GetAccessToken(context.Background(), "garbage")

	s.Require().ErrorIs(err, auth.ErrInvalidToken)
}

func (s *AuthServiceSuite) TestGetAccessToken_RejectsAccessTokenAsRefresh() {
	pair := s.jwtLogin()

	// access-токен подписан другим секретом, поэтому как refresh он невалиден
	_, err := s.service.GetAccessToken(context.Background(), pair.AccessToken)

	s.Require().ErrorIs(err, auth.ErrInvalidToken)
}

func (s *AuthServiceSuite) TestGetAccessToken_ExpiredRefreshToken() {
	repo := mocks.NewIAMRepository(s.T())
	expired := auth.NewService(repo, fakeJWTConfig{accessTTL: time.Minute, refreshTTL: -time.Minute})
	repo.On("GetUserByLogin", mock.Anything, testLogin).Return(s.newUser(), nil)
	pair, err := expired.JWTLogin(context.Background(), testLogin, testPassword)
	s.Require().NoError(err)

	_, err = expired.GetAccessToken(context.Background(), pair.RefreshToken)

	s.Require().ErrorIs(err, auth.ErrInvalidToken)
}

func (s *AuthServiceSuite) TestGetAccessToken_RevokedToken() {
	pair := s.jwtLogin()
	s.repo.On("GetSessionByUUID", mock.Anything, mock.AnythingOfType("uuid.UUID")).
		Return(s.newSession(s.newUser().UUID, time.Hour), nil)

	_, err := s.service.GetAccessToken(context.Background(), pair.RefreshToken)

	s.Require().Error(err)
	s.Contains(err.Error(), "токен отозван")
}

func (s *AuthServiceSuite) TestGetAccessToken_UserNotFound() {
	pair := s.jwtLogin()
	s.expectNotRevoked()
	s.repo.On("GetUserByUUID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(nil, model.ErrUserNotFound)

	_, err := s.service.GetAccessToken(context.Background(), pair.RefreshToken)

	s.Require().ErrorIs(err, model.ErrUserNotFound)
}

func (s *AuthServiceSuite) TestGetAccessToken_Success() {
	pair := s.jwtLogin()
	u := s.newUser()
	s.expectNotRevoked()
	s.repo.On("GetUserByUUID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(u, nil)

	got, err := s.service.GetAccessToken(context.Background(), pair.RefreshToken)

	s.Require().NoError(err)
	s.NotEmpty(got.AccessToken)
	s.Equal(pair.RefreshToken, got.RefreshToken, "refresh-токен должен вернуться без изменений")
	s.True(got.RefreshTokenExpiresAt.IsZero())
}

func (s *AuthServiceSuite) TestGetRefreshToken_EmptyToken() {
	_, err := s.service.GetRefreshToken(context.Background(), "")

	s.Require().ErrorIs(err, auth.ErrInvalidToken)
}

func (s *AuthServiceSuite) TestGetRefreshToken_InvalidToken() {
	_, err := s.service.GetRefreshToken(context.Background(), "garbage")

	s.Require().ErrorIs(err, auth.ErrInvalidToken)
}

func (s *AuthServiceSuite) TestGetRefreshToken_UserNotFound() {
	pair := s.jwtLogin()
	s.expectNotRevoked()
	s.repo.On("GetUserByUUID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(nil, errors.New("db down"))

	_, err := s.service.GetRefreshToken(context.Background(), pair.RefreshToken)

	s.Require().Error(err)
}

func (s *AuthServiceSuite) TestGetRefreshToken_Success() {
	pair := s.jwtLogin()
	s.expectNotRevoked()
	s.repo.On("GetUserByUUID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(s.newUser(), nil)

	got, err := s.service.GetRefreshToken(context.Background(), pair.RefreshToken)

	s.Require().NoError(err)
	s.NotEmpty(got.AccessToken)
	s.NotEmpty(got.RefreshToken)
	s.False(got.RefreshTokenExpiresAt.IsZero())
}
