package health

import (
	"context"
	"errors"
	"testing"
	"tiny-bitly/internal/dao"
	mock_dao "tiny-bitly/internal/dao/generated"
	"tiny-bitly/internal/model"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

type HealthServiceSuite struct {
	suite.Suite
	ctrl         *gomock.Controller
	service      *Service
	dao          dao.DAO
	urlRecordDAO *mock_dao.MockURLRecordDAO
}

func TestHealthServiceSuite(t *testing.T) {
	suite.Run(t, new(HealthServiceSuite))
}

func (suite *HealthServiceSuite) SetupTest() {
	suite.ctrl = gomock.NewController(suite.T())
	suite.urlRecordDAO = mock_dao.NewMockURLRecordDAO(suite.ctrl)
	suite.dao = dao.DAO{
		URLRecordDAO: suite.urlRecordDAO,
	}
	suite.service = NewService(suite.dao)
}

func (suite *HealthServiceSuite) TestCheckHealthSuccess() {
	// Mock: DAO responds successfully and record does not exist.
	suite.urlRecordDAO.
		EXPECT().
		GetByShortCode(gomock.Any(), "__health_check__").
		Return(nil, nil)

	isHealthy := suite.service.CheckHealth(context.Background())
	suite.True(isHealthy)
}

func (suite *HealthServiceSuite) TestCheckHealthWithRecordFound() {
	// Mock: DAO responds successfully and record exists.
	suite.urlRecordDAO.
		EXPECT().
		GetByShortCode(gomock.Any(), "__health_check__").
		Return(&model.URLRecordEntity{}, nil)

	isHealthy := suite.service.CheckHealth(context.Background())
	suite.True(isHealthy)
}

func (suite *HealthServiceSuite) TestCheckHealthFailure() {
	// Mock: DAO fails to respond.
	suite.urlRecordDAO.
		EXPECT().
		GetByShortCode(gomock.Any(), "__health_check__").
		Return(nil, errors.New("database connection failed"))

	isHealthy := suite.service.CheckHealth(context.Background())
	suite.False(isHealthy)
}
