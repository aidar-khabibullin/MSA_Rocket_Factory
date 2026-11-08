package auth_test

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/aidar-khabibullin/MSA_Rocket_Factory/iam/internal/model"
	"github.com/aidar-khabibullin/MSA_Rocket_Factory/iam/internal/service/auth"
)

func (s *AuthServiceSuite) newJWTService() *auth.JWTService {
	return auth.NewJWTService(s.jwtConfig, auth.NewTokenBlacklistService(s.repo))
}

// signClaims подписывает произвольные claims указанным секретом и алгоритмом
func (s *AuthServiceSuite) signClaims(method jwt.SigningMethod, secret string, claims jwt.MapClaims) string {
	str, err := jwt.NewWithClaims(method, claims).SignedString([]byte(secret))
	s.Require().NoError(err)

	return str
}

func (s *AuthServiceSuite) validAccessClaims() jwt.MapClaims {
	return jwt.MapClaims{
		"type":      "access",
		"user_id":   float64(7),
		"username":  testLogin,
		"user_uuid": uuid.NewString(),
		"exp":       time.Now().Add(time.Hour).Unix(),
	}
}

func (s *AuthServiceSuite) TestValidateAccessToken_Success() {
	s.expectNotRevoked()
	pair := s.jwtLogin()

	claims, err := s.newJWTService().ValidateAccessToken(pair.AccessToken)

	s.Require().NoError(err)
	s.Equal(testLogin, claims.Username)
	s.NotEqual(uuid.Nil, claims.UserUUID)
}

func (s *AuthServiceSuite) TestValidateAccessToken_Revoked() {
	pair := s.jwtLogin()
	s.repo.On("GetSessionByUUID", mock.Anything, mock.AnythingOfType("uuid.UUID")).
		Return(&model.Session{}, nil)

	_, err := s.newJWTService().ValidateAccessTokenWithContext(context.Background(), pair.AccessToken)

	s.Require().Error(err)
	s.NotErrorIs(err, auth.ErrInvalidToken)
}

func (s *AuthServiceSuite) TestValidateAccessToken_RefreshTokenIsRejected() {
	pair := s.jwtLogin()

	// refresh-токен подписан другим секретом, поэтому подпись не сойдётся
	_, err := s.newJWTService().ValidateAccessToken(pair.RefreshToken)

	s.Require().ErrorIs(err, auth.ErrInvalidToken)
}

func (s *AuthServiceSuite) expectNotRevokedMaybe() {
	s.repo.On("GetSessionByUUID", mock.Anything, mock.AnythingOfType("uuid.UUID")).
		Return(nil, model.ErrSessionNotFound).Maybe()
}

func (s *AuthServiceSuite) TestValidateAccessToken_InvalidTokens() {
	s.expectNotRevokedMaybe()

	missing := func(key string) jwt.MapClaims {
		c := s.validAccessClaims()
		delete(c, key)

		return c
	}
	withValue := func(key string, value any) jwt.MapClaims {
		c := s.validAccessClaims()
		c[key] = value

		return c
	}

	const secret = "test-access-secret"
	cases := map[string]string{
		"garbage":             "not-a-jwt",
		"wrong secret":        s.signClaims(jwt.SigningMethodHS256, "other-secret", s.validAccessClaims()),
		"wrong algorithm":     s.signClaims(jwt.SigningMethodHS512, secret, s.validAccessClaims()),
		"expired":             s.signClaims(jwt.SigningMethodHS256, secret, withValue("exp", time.Now().Add(-time.Hour).Unix())),
		"wrong token type":    s.signClaims(jwt.SigningMethodHS256, secret, withValue("type", "refresh")),
		"no token type":       s.signClaims(jwt.SigningMethodHS256, secret, missing("type")),
		"no user_id":          s.signClaims(jwt.SigningMethodHS256, secret, missing("user_id")),
		"no username":         s.signClaims(jwt.SigningMethodHS256, secret, missing("username")),
		"no user_uuid":        s.signClaims(jwt.SigningMethodHS256, secret, missing("user_uuid")),
		"malformed user_uuid": s.signClaims(jwt.SigningMethodHS256, secret, withValue("user_uuid", "not-a-uuid")),
		"user_id not number":  s.signClaims(jwt.SigningMethodHS256, secret, withValue("user_id", "seven")),
	}

	for name, token := range cases {
		s.Run(name, func() {
			_, err := s.newJWTService().ValidateAccessToken(token)

			s.Require().ErrorIs(err, auth.ErrInvalidToken)
		})
	}
}

func (s *AuthServiceSuite) TestValidateRefreshToken_ViaGetAccessToken_InvalidClaims() {
	s.expectNotRevokedMaybe()

	const secret = "test-refresh-secret"
	base := func() jwt.MapClaims {
		return jwt.MapClaims{
			"type":      "refresh",
			"user_id":   float64(7),
			"username":  testLogin,
			"user_uuid": uuid.NewString(),
			"exp":       time.Now().Add(time.Hour).Unix(),
		}
	}
	without := func(key string) jwt.MapClaims {
		c := base()
		delete(c, key)

		return c
	}
	badUUID := base()
	badUUID["user_uuid"] = "not-a-uuid"
	wrongType := base()
	wrongType["type"] = "access"

	cases := map[string]string{
		"wrong algorithm": s.signClaims(jwt.SigningMethodHS512, secret, base()),
		"wrong type":      s.signClaims(jwt.SigningMethodHS256, secret, wrongType),
		"no user_id":      s.signClaims(jwt.SigningMethodHS256, secret, without("user_id")),
		"no username":     s.signClaims(jwt.SigningMethodHS256, secret, without("username")),
		"no user_uuid":    s.signClaims(jwt.SigningMethodHS256, secret, without("user_uuid")),
		"bad user_uuid":   s.signClaims(jwt.SigningMethodHS256, secret, badUUID),
	}

	for name, token := range cases {
		s.Run(name, func() {
			_, err := s.service.GetAccessToken(context.Background(), token)

			s.Require().Error(err)
		})
	}
}

func (s *AuthServiceSuite) TestRevokeAllUserTokens_StoresMarker() {
	userUUID := uuid.New()
	s.repo.On("Set", mock.Anything, mock.AnythingOfType("uuid.UUID"),
		mock.MatchedBy(func(sess *model.Session) bool { return sess.UserUUID == userUUID }),
		48*time.Hour).Return(nil)

	blacklist := auth.NewTokenBlacklistService(s.repo)

	s.Require().NoError(blacklist.RevokeAllUserTokens(context.Background(), userUUID))
	s.False(blacklist.IsUserTokensRevoked(context.Background(), userUUID, time.Now()))
}

func (s *AuthServiceSuite) TestRevokeAllUserTokens_CacheError() {
	cacheErr := errors.New("redis down")
	s.repo.On("Set", mock.Anything, mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("*model.Session"), mock.Anything).
		Return(cacheErr)

	err := auth.NewTokenBlacklistService(s.repo).RevokeAllUserTokens(context.Background(), uuid.New())

	s.Require().ErrorIs(err, cacheErr)
}
