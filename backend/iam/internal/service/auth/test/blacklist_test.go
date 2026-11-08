package auth_test

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/aidar-khabibullin/MSA_Rocket_Factory/iam/internal/model"
	"github.com/aidar-khabibullin/MSA_Rocket_Factory/iam/internal/repository/mocks"
	"github.com/aidar-khabibullin/MSA_Rocket_Factory/iam/internal/service/auth"
)

func signedToken(s *AuthServiceSuite, exp time.Time) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"exp": exp.Unix()})
	str, err := token.SignedString([]byte("any-secret"))
	s.Require().NoError(err)

	return str
}

func (s *AuthServiceSuite) TestRevokeToken_StoresEntryWithTTL() {
	token := signedToken(s, time.Now().Add(time.Hour))
	s.repo.On("Set", mock.Anything, mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("*model.Session"), mock.MatchedBy(func(ttl time.Duration) bool {
		return ttl > 59*time.Minute && ttl <= time.Hour
	})).Return(nil)

	s.Require().NoError(s.service.RevokeToken(context.Background(), token))
}

func (s *AuthServiceSuite) TestRevokeToken_AlreadyExpiredIsNoop() {
	token := signedToken(s, time.Now().Add(-time.Hour))

	// Set не ожидается: мок упадёт, если сервис попытается что-то записать
	s.Require().NoError(s.service.RevokeToken(context.Background(), token))
}

func (s *AuthServiceSuite) TestRevokeToken_MalformedToken() {
	s.Require().Error(s.service.RevokeToken(context.Background(), "not-a-jwt"))
}

func (s *AuthServiceSuite) TestRevokeToken_TokenWithoutExp() {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": "x"})
	str, err := token.SignedString([]byte("any-secret"))
	s.Require().NoError(err)

	s.Require().Error(s.service.RevokeToken(context.Background(), str))
}

func (s *AuthServiceSuite) TestRevokeToken_CacheError() {
	token := signedToken(s, time.Now().Add(time.Hour))
	s.repo.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errors.New("redis down"))

	s.Require().Error(s.service.RevokeToken(context.Background(), token))
}

// Регрессия: ключ blacklist раньше строился через uuid.MustParse от 64-символьного
// SHA-256, что приводило к panic при любой проверке токена.
func (s *AuthServiceSuite) TestBlacklist_KeyIsStableAndDoesNotPanic() {
	repo := mocks.NewIAMRepository(s.T())
	bl := auth.NewTokenBlacklistService(repo)
	token := signedToken(s, time.Now().Add(time.Hour))

	var setKey, getKey uuid.UUID
	repo.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) { setKey = args.Get(1).(uuid.UUID) }).Return(nil)
	repo.On("GetSessionByUUID", mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) { getKey = args.Get(1).(uuid.UUID) }).
		Return(&model.Session{}, nil)

	s.Require().NoError(bl.RevokeToken(context.Background(), token))
	s.Require().NotPanics(func() {
		s.True(bl.IsTokenRevoked(context.Background(), token))
	})
	s.Equal(setKey, getKey, "один и тот же токен должен давать один и тот же ключ")
}

func (s *AuthServiceSuite) TestBlacklist_IsTokenRevoked_NotFound() {
	repo := mocks.NewIAMRepository(s.T())
	repo.On("GetSessionByUUID", mock.Anything, mock.Anything).Return(nil, model.ErrSessionNotFound)

	s.False(auth.NewTokenBlacklistService(repo).IsTokenRevoked(context.Background(), "tok"))
}

func (s *AuthServiceSuite) TestBlacklist_RevokeAllUserTokens() {
	repo := mocks.NewIAMRepository(s.T())
	bl := auth.NewTokenBlacklistService(repo)
	userUUID := uuid.New()

	repo.On("Set", mock.Anything, mock.Anything, mock.MatchedBy(func(sess *model.Session) bool {
		return sess.UserUUID == userUUID
	}), 48*time.Hour).Return(nil).Once()
	s.Require().NoError(bl.RevokeAllUserTokens(context.Background(), userUUID))

	repo.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errors.New("redis down")).Once()
	s.Require().Error(bl.RevokeAllUserTokens(context.Background(), userUUID))
}
