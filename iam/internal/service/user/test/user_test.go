package user_test

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/aidar-khabibullin/MSA_Rocket_Factory/iam/internal/model"
)

func validInfo() *model.UserRegistrationInfo {
	return &model.UserRegistrationInfo{
		Login:    "rocket_user",
		Email:    "rocket@example.com",
		Password: "Str0ng!Passw0rd",
		NotificationMethods: []model.NotificationMethod{
			{ProviderName: "telegram", Target: "12345"},
		},
	}
}

func (s *UserServiceSuite) TestRegister_ValidationErrors() {
	cases := map[string]func(*model.UserRegistrationInfo){
		"short login":   func(i *model.UserRegistrationInfo) { i.Login = "ab" },
		"bad login":     func(i *model.UserRegistrationInfo) { i.Login = "bad login!" },
		"empty email":   func(i *model.UserRegistrationInfo) { i.Email = "" },
		"bad email":     func(i *model.UserRegistrationInfo) { i.Email = "not-an-email" },
		"short pass":    func(i *model.UserRegistrationInfo) { i.Password = "Aa1!" },
		"no upper":      func(i *model.UserRegistrationInfo) { i.Password = "str0ng!passw0rd" },
		"no digit":      func(i *model.UserRegistrationInfo) { i.Password = "Strong!Password" },
		"no special":    func(i *model.UserRegistrationInfo) { i.Password = "Str0ngPassw0rd" },
		"too long pass": func(i *model.UserRegistrationInfo) { i.Password = "Aa1!" + strings.Repeat("a", 200) },
	}

	for name, mutate := range cases {
		s.Run(name, func() {
			info := validInfo()
			mutate(info)

			got, err := s.service.Register(context.Background(), info)

			s.Require().Error(err)
			s.Nil(got)
			// В репозиторий при невалидных данных ходить нельзя (мок упал бы на вызове)
		})
	}
}

func (s *UserServiceSuite) TestRegister_UserAlreadyExists() {
	info := validInfo()
	s.repo.On("GetUserByLogin", mock.Anything, info.Login).Return(&model.User{Login: info.Login}, nil)

	got, err := s.service.Register(context.Background(), info)

	s.Require().ErrorIs(err, model.ErrUserAlreadyExists)
	s.Nil(got)
}

func (s *UserServiceSuite) TestRegister_CreateUserFails() {
	info := validInfo()
	dbErr := errors.New("db down")
	s.repo.On("GetUserByLogin", mock.Anything, info.Login).Return(nil, model.ErrUserNotFound)
	s.repo.On("CreateUser", mock.Anything, mock.AnythingOfType("*model.User")).Return(nil, dbErr)

	got, err := s.service.Register(context.Background(), info)

	s.Require().ErrorIs(err, dbErr)
	s.Nil(got)
}

func (s *UserServiceSuite) TestRegister_Success_HashesPassword() {
	info := validInfo()
	var saved *model.User
	s.repo.On("GetUserByLogin", mock.Anything, info.Login).Return(nil, model.ErrUserNotFound)
	s.repo.On("CreateUser", mock.Anything, mock.AnythingOfType("*model.User")).
		Run(func(args mock.Arguments) { saved = args.Get(1).(*model.User) }).
		Return(func(_ context.Context, u *model.User) *model.User { return u }, nil)

	got, err := s.service.Register(context.Background(), info)

	s.Require().NoError(err)
	s.Require().NotNil(got)
	s.Equal(info.Login, got.Login)
	s.Equal(info.Email, got.Email)
	s.Equal(info.NotificationMethods, got.NotificationMethods)
	s.NotEqual(uuid.Nil, got.UUID)

	s.Require().NotNil(saved)
	s.NotContains(saved.PasswordHash, info.Password, "пароль не должен храниться в открытом виде")
	s.True(strings.HasPrefix(saved.PasswordHash, "$argon2id$v=19$m=65536,t=1,p=4$"))
}

func (s *UserServiceSuite) TestRegister_SamePasswordGivesDifferentHashes() {
	hashes := make([]string, 0, 2)
	s.repo.On("GetUserByLogin", mock.Anything, mock.Anything).Return(nil, model.ErrUserNotFound)
	s.repo.On("CreateUser", mock.Anything, mock.AnythingOfType("*model.User")).
		Run(func(args mock.Arguments) { hashes = append(hashes, args.Get(1).(*model.User).PasswordHash) }).
		Return(func(_ context.Context, u *model.User) *model.User { return u }, nil)

	for range 2 {
		_, err := s.service.Register(context.Background(), validInfo())
		s.Require().NoError(err)
	}

	s.Require().Len(hashes, 2)
	s.NotEqual(hashes[0], hashes[1], "соль должна быть уникальной")
}

func (s *UserServiceSuite) TestGetUser() {
	id := uuid.New()
	want := &model.User{UUID: id, Login: "rocket_user"}
	s.repo.On("GetUserByUUID", mock.Anything, id).Return(want, nil).Once()

	got, err := s.service.GetUser(context.Background(), id)
	s.Require().NoError(err)
	s.Equal(want, got)

	s.repo.On("GetUserByUUID", mock.Anything, id).Return(nil, model.ErrUserNotFound).Once()
	_, err = s.service.GetUser(context.Background(), id)
	s.Require().ErrorIs(err, model.ErrUserNotFound)
}
