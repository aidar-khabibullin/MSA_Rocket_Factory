package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/aidar-khabibullin/MSA_Rocket_Factory/iam/internal/model"
	"github.com/aidar-khabibullin/MSA_Rocket_Factory/iam/internal/repository/mocks"
	"github.com/aidar-khabibullin/MSA_Rocket_Factory/iam/internal/service/auth"
	"github.com/aidar-khabibullin/MSA_Rocket_Factory/iam/internal/service/user"
	"github.com/aidar-khabibullin/MSA_Rocket_Factory/platform/pkg/logger"
)

const (
	testPassword = "Str0ng!Passw0rd"
	testLogin    = "rocket_user"
)

// authService описывает публичные методы сервиса, которые проверяются в тестах
type authService interface {
	Login(ctx context.Context, login, password string) (*model.Session, error)
	Whoami(ctx context.Context, sessionUUID uuid.UUID) (*model.Session, *model.User, error)
	Logout(ctx context.Context, sessionUUID uuid.UUID) error
	JWTLogin(ctx context.Context, login, password string) (*model.TokenPair, error)
	GetAccessToken(ctx context.Context, refreshToken string) (*model.TokenPair, error)
	GetRefreshToken(ctx context.Context, refreshToken string) (*model.TokenPair, error)
	RevokeToken(ctx context.Context, tokenString string) error
}

// fakeJWTConfig - тестовая реализация env.JWTConfig
type fakeJWTConfig struct {
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func (c fakeJWTConfig) AccessTokenSecret() string  { return "test-access-secret" }
func (c fakeJWTConfig) RefreshTokenSecret() string { return "test-refresh-secret" }
func (c fakeJWTConfig) AccessTokenTTL() time.Duration {
	return c.accessTTL
}

func (c fakeJWTConfig) RefreshTokenTTL() time.Duration {
	return c.refreshTTL
}

type AuthServiceSuite struct {
	suite.Suite

	repo    *mocks.IAMRepository
	service authService
	// internalService дает доступ к методам, не входящим в интерфейс def.AuthService
	jwtConfig fakeJWTConfig

	// passwordHash - валидный Argon2-хэш для testPassword (считается один раз)
	passwordHash string
}

func (s *AuthServiceSuite) SetupSuite() {
	logger.SetNopLogger()
	s.jwtConfig = fakeJWTConfig{accessTTL: 15 * time.Minute, refreshTTL: 24 * time.Hour}
	s.passwordHash = s.buildPasswordHash(testPassword)
}

func (s *AuthServiceSuite) SetupTest() {
	s.repo = mocks.NewIAMRepository(s.T())
	s.service = auth.NewService(s.repo, s.jwtConfig)
}

// buildPasswordHash получает настоящий хэш через регистрацию пользователя
func (s *AuthServiceSuite) buildPasswordHash(password string) string {
	repo := mocks.NewIAMRepository(s.T())
	repo.On("GetUserByLogin", mock.Anything, testLogin).Return(nil, model.ErrUserNotFound)

	var hash string
	repo.On("CreateUser", mock.Anything, mock.AnythingOfType("*model.User")).
		Run(func(args mock.Arguments) {
			hash = args.Get(1).(*model.User).PasswordHash
		}).
		Return(func(_ context.Context, u *model.User) *model.User { return u }, nil)

	_, err := user.NewService(repo).Register(context.Background(), &model.UserRegistrationInfo{
		Login:    testLogin,
		Email:    "rocket@example.com",
		Password: password,
	})
	s.Require().NoError(err)
	s.Require().NotEmpty(hash)

	return hash
}

func (s *AuthServiceSuite) newUser() *model.User {
	return &model.User{
		UUID:         mustUUID(),
		Login:        testLogin,
		Email:        "rocket@example.com",
		PasswordHash: s.passwordHash,
	}
}

func TestAuthServiceSuite(t *testing.T) {
	suite.Run(t, new(AuthServiceSuite))
}
