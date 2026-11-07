package auth_test

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/aidar-khabibullin/MSA_Rocket_Factory/iam/internal/model"
)

func (s *AuthServiceSuite) newSession(userUUID uuid.UUID, expiresIn time.Duration) *model.Session {
	return &model.Session{
		UUID:      uuid.New(),
		UserUUID:  userUUID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		ExpiresAt: time.Now().Add(expiresIn),
	}
}

func (s *AuthServiceSuite) TestWhoami_CacheHit() {
	u := s.newUser()
	sess := s.newSession(u.UUID, time.Hour)
	s.repo.On("GetSessionFromCache", mock.Anything, sess.UUID).Return(sess, nil)
	s.repo.On("GetUserByUUID", mock.Anything, u.UUID).Return(u, nil)

	gotSession, gotUser, err := s.service.Whoami(context.Background(), sess.UUID)

	s.Require().NoError(err)
	s.Equal(sess, gotSession)
	s.Equal(u, gotUser)
	// В БД за сессией ходить не должны: мок упадёт на неожиданном вызове GetSessionByUUID
}

func (s *AuthServiceSuite) TestWhoami_CacheMiss_LoadsFromDBAndCaches() {
	u := s.newUser()
	sess := s.newSession(u.UUID, time.Hour)
	s.repo.On("GetSessionFromCache", mock.Anything, sess.UUID).Return(nil, model.ErrSessionNotFound)
	s.repo.On("GetSessionByUUID", mock.Anything, sess.UUID).Return(sess, nil)
	s.repo.On("Set", mock.Anything, sess.UUID, sess, mock.AnythingOfType("time.Duration")).Return(nil)
	s.repo.On("GetUserByUUID", mock.Anything, u.UUID).Return(u, nil)

	gotSession, gotUser, err := s.service.Whoami(context.Background(), sess.UUID)

	s.Require().NoError(err)
	s.Equal(sess, gotSession)
	s.Equal(u, gotUser)
}

func (s *AuthServiceSuite) TestWhoami_CacheError_FallsBackToDB() {
	u := s.newUser()
	sess := s.newSession(u.UUID, time.Hour)
	s.repo.On("GetSessionFromCache", mock.Anything, sess.UUID).Return(nil, errors.New("redis down"))
	s.repo.On("GetSessionByUUID", mock.Anything, sess.UUID).Return(sess, nil)
	s.repo.On("Set", mock.Anything, sess.UUID, sess, mock.Anything).Return(errors.New("redis down"))
	s.repo.On("GetUserByUUID", mock.Anything, u.UUID).Return(u, nil)

	_, gotUser, err := s.service.Whoami(context.Background(), sess.UUID)

	s.Require().NoError(err)
	s.Equal(u, gotUser)
}

func (s *AuthServiceSuite) TestWhoami_SessionNotFoundInDB() {
	id := uuid.New()
	s.repo.On("GetSessionFromCache", mock.Anything, id).Return(nil, model.ErrSessionNotFound)
	s.repo.On("GetSessionByUUID", mock.Anything, id).Return(nil, model.ErrSessionNotFound)

	session, user, err := s.service.Whoami(context.Background(), id)

	s.Require().ErrorIs(err, model.ErrSessionNotFound)
	s.Nil(session)
	s.Nil(user)
}

func (s *AuthServiceSuite) TestWhoami_ExpiredSession_IsEvictedFromCache() {
	sess := s.newSession(uuid.New(), -time.Minute)
	s.repo.On("GetSessionFromCache", mock.Anything, sess.UUID).Return(sess, nil)
	s.repo.On("Delete", mock.Anything, sess.UUID).Return(nil)

	session, user, err := s.service.Whoami(context.Background(), sess.UUID)

	s.Require().ErrorIs(err, model.ErrSessionExpired)
	s.Nil(session)
	s.Nil(user)
}

func (s *AuthServiceSuite) TestWhoami_UserLookupFails() {
	sess := s.newSession(uuid.New(), time.Hour)
	s.repo.On("GetSessionFromCache", mock.Anything, sess.UUID).Return(sess, nil)
	s.repo.On("GetUserByUUID", mock.Anything, sess.UserUUID).Return(nil, model.ErrUserNotFound)

	_, _, err := s.service.Whoami(context.Background(), sess.UUID)

	s.Require().ErrorIs(err, model.ErrUserNotFound)
}

func (s *AuthServiceSuite) TestLogout_Success() {
	id := uuid.New()
	s.repo.On("DeleteSession", mock.Anything, id).Return(nil)
	s.repo.On("Delete", mock.Anything, id).Return(nil)

	s.Require().NoError(s.service.Logout(context.Background(), id))
}

func (s *AuthServiceSuite) TestLogout_DBError() {
	id := uuid.New()
	dbErr := errors.New("db down")
	s.repo.On("DeleteSession", mock.Anything, id).Return(dbErr)

	s.Require().ErrorIs(s.service.Logout(context.Background(), id), dbErr)
}

func (s *AuthServiceSuite) TestLogout_CacheErrorIsIgnored() {
	id := uuid.New()
	s.repo.On("DeleteSession", mock.Anything, id).Return(nil)
	s.repo.On("Delete", mock.Anything, id).Return(errors.New("redis down"))

	s.Require().NoError(s.service.Logout(context.Background(), id))
}
