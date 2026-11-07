package auth_test

import (
	"context"
	"errors"
	"time"

	"github.com/stretchr/testify/mock"

	"github.com/aidar-khabibullin/MSA_Rocket_Factory/iam/internal/model"
)

func (s *AuthServiceSuite) TestLogin_EmptyLogin() {
	session, err := s.service.Login(context.Background(), "", testPassword)

	s.Require().ErrorIs(err, model.ErrEmptyLogin)
	s.Nil(session)
}

func (s *AuthServiceSuite) TestLogin_EmptyPassword() {
	session, err := s.service.Login(context.Background(), testLogin, "")

	s.Require().ErrorIs(err, model.ErrEmptyPassword)
	s.Nil(session)
}

func (s *AuthServiceSuite) TestLogin_UserNotFound() {
	s.repo.On("GetUserByLogin", mock.Anything, testLogin).Return(nil, model.ErrUserNotFound)

	session, err := s.service.Login(context.Background(), testLogin, testPassword)

	// Наружу не раскрываем, что пользователя нет
	s.Require().ErrorIs(err, model.ErrInvalidCredentials)
	s.Nil(session)
}

func (s *AuthServiceSuite) TestLogin_WrongPassword() {
	s.repo.On("GetUserByLogin", mock.Anything, testLogin).Return(s.newUser(), nil)

	session, err := s.service.Login(context.Background(), testLogin, "Wrong!Passw0rd")

	s.Require().ErrorIs(err, model.ErrInvalidCredentials)
	s.Nil(session)
}

func (s *AuthServiceSuite) TestLogin_CorruptedHash() {
	u := s.newUser()
	u.PasswordHash = "not-a-valid-hash"
	s.repo.On("GetUserByLogin", mock.Anything, testLogin).Return(u, nil)

	session, err := s.service.Login(context.Background(), testLogin, testPassword)

	s.Require().ErrorIs(err, model.ErrPasswordVerification)
	s.Nil(session)
}

func (s *AuthServiceSuite) TestLogin_CreateSessionFails() {
	s.repo.On("GetUserByLogin", mock.Anything, testLogin).Return(s.newUser(), nil)
	s.repo.On("CreateSession", mock.Anything, mock.AnythingOfType("*model.Session")).
		Return(nil, errors.New("db down"))

	session, err := s.service.Login(context.Background(), testLogin, testPassword)

	s.Require().ErrorIs(err, model.ErrFailedToCreateSession)
	s.Nil(session)
}

func (s *AuthServiceSuite) TestLogin_Success() {
	u := s.newUser()
	s.repo.On("GetUserByLogin", mock.Anything, testLogin).Return(u, nil)
	s.repo.On("CreateSession", mock.Anything, mock.AnythingOfType("*model.Session")).
		Return(func(_ context.Context, sess *model.Session) *model.Session { return sess }, nil)
	s.repo.On("Set", mock.Anything, mock.Anything, mock.AnythingOfType("*model.Session"), 24*time.Hour).Return(nil)

	session, err := s.service.Login(context.Background(), testLogin, testPassword)

	s.Require().NoError(err)
	s.Require().NotNil(session)
	s.Equal(u.UUID, session.UserUUID)
	s.True(session.ExpiresAt.After(time.Now()))
}

func (s *AuthServiceSuite) TestLogin_CacheFailureDoesNotFailLogin() {
	s.repo.On("GetUserByLogin", mock.Anything, testLogin).Return(s.newUser(), nil)
	s.repo.On("CreateSession", mock.Anything, mock.AnythingOfType("*model.Session")).
		Return(func(_ context.Context, sess *model.Session) *model.Session { return sess }, nil)
	s.repo.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errors.New("redis down"))

	session, err := s.service.Login(context.Background(), testLogin, testPassword)

	s.Require().NoError(err)
	s.NotNil(session)
}

// Регрессия: ответ для несуществующего логина не должен быть заметно быстрее ответа
// для неверного пароля, иначе по времени можно перебирать существующие логины.
func (s *AuthServiceSuite) TestLogin_UnknownUserTakesComparableTimeToWrongPassword() {
	const attempts = 3

	measure := func(mockUser *model.User, mockErr error) time.Duration {
		s.repo.ExpectedCalls = nil
		s.repo.On("GetUserByLogin", mock.Anything, testLogin).Return(mockUser, mockErr)

		start := time.Now()
		for range attempts {
			_, err := s.service.Login(context.Background(), testLogin, "Wrong!Passw0rd")
			s.Require().ErrorIs(err, model.ErrInvalidCredentials)
		}

		return time.Since(start) / attempts
	}

	wrongPassword := measure(s.newUser(), nil)
	unknownUser := measure(nil, model.ErrUserNotFound)

	// Argon2 (64 МБ) занимает десятки миллисекунд, а обращение к мокам - микросекунды,
	// поэтому порог 50% надёжно отличает "хэш посчитан" от "хэш пропущен".
	s.Greater(unknownUser, wrongPassword/2,
		"unknown user: %s, wrong password: %s", unknownUser, wrongPassword)
}

func (s *AuthServiceSuite) TestJWTLogin_UnknownUserTakesComparableTimeToWrongPassword() {
	const attempts = 3

	measure := func(mockUser *model.User, mockErr error) time.Duration {
		s.repo.ExpectedCalls = nil
		s.repo.On("GetUserByLogin", mock.Anything, testLogin).Return(mockUser, mockErr)

		start := time.Now()
		for range attempts {
			_, err := s.service.JWTLogin(context.Background(), testLogin, "Wrong!Passw0rd")
			s.Require().ErrorIs(err, model.ErrInvalidCredentials)
		}

		return time.Since(start) / attempts
	}

	wrongPassword := measure(s.newUser(), nil)
	unknownUser := measure(nil, model.ErrUserNotFound)

	s.Greater(unknownUser, wrongPassword/2,
		"unknown user: %s, wrong password: %s", unknownUser, wrongPassword)
}
