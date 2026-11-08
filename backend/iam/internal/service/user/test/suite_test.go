package user_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/aidar-khabibullin/MSA_Rocket_Factory/iam/internal/repository/mocks"
	"github.com/aidar-khabibullin/MSA_Rocket_Factory/iam/internal/service"
	"github.com/aidar-khabibullin/MSA_Rocket_Factory/iam/internal/service/user"
	"github.com/aidar-khabibullin/MSA_Rocket_Factory/platform/pkg/logger"
)

type UserServiceSuite struct {
	suite.Suite

	repo    *mocks.IAMRepository
	service service.UserService
}

func (s *UserServiceSuite) SetupSuite() {
	logger.SetNopLogger()
}

func (s *UserServiceSuite) SetupTest() {
	s.repo = mocks.NewIAMRepository(s.T())
	s.service = user.NewService(s.repo)
}

func TestUserServiceSuite(t *testing.T) {
	suite.Run(t, new(UserServiceSuite))
}
