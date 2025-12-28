package create

import (
	"context"
	"strings"
	"testing"
	"tiny-bitly/internal/apperrors"
	"tiny-bitly/internal/config"
	"tiny-bitly/internal/dao"
	mock_daotypes "tiny-bitly/internal/dao/generated"
	"tiny-bitly/internal/model"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

type CreateServiceSuite struct {
	suite.Suite
	ctx          context.Context
	ctrl         *gomock.Controller
	service      *Service
	dao          dao.DAO
	urlRecordDAO *mock_daotypes.MockURLRecordDAO
}

func TestCreateServiceSuite(t *testing.T) {
	suite.Run(t, new(CreateServiceSuite))
}

func (suite *CreateServiceSuite) SetupTest() {
	suite.ctx = context.Background()
	suite.ctrl = gomock.NewController(suite.T())
	suite.urlRecordDAO = mock_daotypes.NewMockURLRecordDAO(suite.ctrl)
	suite.dao = dao.DAO{
		URLRecordDAO: suite.urlRecordDAO,
	}
	cfg := config.GetTestConfig(config.Config{})
	suite.service = NewService(suite.dao, &cfg)
}

func (suite *CreateServiceSuite) TestErrorInputURLEmpty() {
	originalURL := ""
	_, err := suite.service.CreateShortURL(suite.ctx, originalURL, nil)
	suite.NotNil(err)
	suite.ErrorContains(err, "invalid URL")
}

func (suite *CreateServiceSuite) TestErrorInputURLInvalidChars() {
	originalURL := "www.`.com"
	_, err := suite.service.CreateShortURL(suite.ctx, originalURL, nil)
	suite.NotNil(err)
	suite.ErrorContains(err, "invalid URL")
}

func (suite *CreateServiceSuite) TestErrorInputURLTooLong() {
	cfg := config.GetTestConfig(config.Config{MaxURLLength: 2})
	service := NewService(suite.dao, &cfg)
	originalURL := "abc"
	_, err := service.CreateShortURL(suite.ctx, originalURL, nil)
	suite.NotNil(err)
	suite.ErrorIs(err, apperrors.ErrURLLengthExceeded)
}

func (suite *CreateServiceSuite) TestErrorInputAliasEmpty() {
	originalURL := "https://www.foo.com"
	alias := ""
	_, err := suite.service.CreateShortURL(suite.ctx, originalURL, &alias)
	suite.NotNil(err)
	suite.ErrorIs(err, apperrors.ErrInvalidAlias)
}

func (suite *CreateServiceSuite) TestErrorInputAliasInvalidChars() {
	originalURL := "https://www.foo.com"
	alias := "`"
	_, err := suite.service.CreateShortURL(suite.ctx, originalURL, &alias)
	suite.NotNil(err)
	suite.ErrorIs(err, apperrors.ErrInvalidAlias)
}

func (suite *CreateServiceSuite) TestErrorInputAliasAlreadyUsedForDifferentURL() {
	// Mock: Create() should return a specific error code.
	suite.urlRecordDAO.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, apperrors.ErrShortCodeAlreadyInUse)

	originalURL := "https://www.foo.com"
	alias := "bar"
	_, err := suite.service.CreateShortURL(suite.ctx, originalURL, &alias)
	suite.NotNil(err)
	suite.ErrorIs(err, apperrors.ErrAliasAlreadyInUse)
}

func (suite *CreateServiceSuite) TestErrorConfigAPIHostnameMissing() {
	// HACK: Set the API Hostname to empty string.
	cfg := config.GetTestConfig(config.Config{})
	cfg.APIHostname = ""
	service := NewService(suite.dao, &cfg)

	suite.MockCreateSuccess().Times(0)

	originalURL := "https://www.foo.com"
	_, err := service.CreateShortURL(suite.ctx, originalURL, nil)
	suite.NotNil(err)
	suite.ErrorIs(err, apperrors.ErrConfigurationMissing)
}

func (suite *CreateServiceSuite) TestErrorMaxRetries() {
	// Mock: Create() should fail to generate an unused short code.
	suite.MockCreateFail().AnyTimes()

	originalURL := "https://www.foo.com"
	_, err := suite.service.CreateShortURL(suite.ctx, originalURL, nil)
	suite.NotNil(err)
	suite.ErrorIs(err, apperrors.ErrMaxRetriesExceeded)
}

func (suite *CreateServiceSuite) TestConfigMaxTries() {
	// Allow only a single attempt.
	cfg := config.GetTestConfig(config.Config{MaxTriesCreateShortCode: 1})
	service := NewService(suite.dao, &cfg)

	// First call fails.
	suite.MockCreateFail().Times(1)

	// Second call must NEVER happen.
	suite.MockCreateSuccess().Times(0)

	originalURL := "https://www.foo.com"
	_, err := service.CreateShortURL(suite.ctx, originalURL, nil)

	suite.Error(err)
	suite.ErrorIs(err, apperrors.ErrMaxRetriesExceeded)
}

func (suite *CreateServiceSuite) TestConfigShortCodeLength() {
	customShortCodeLength := 8
	cfg := config.GetTestConfig(config.Config{ShortCodeLength: customShortCodeLength})
	service := NewService(suite.dao, &cfg)

	suite.MockCreateSuccess().Times(1)

	originalURL := "https://www.foo.com"
	shortURL, err := service.CreateShortURL(suite.ctx, originalURL, nil)
	suite.Nil(err)
	suite.NotNil(shortURL)
	slashIndex := strings.LastIndex(*shortURL, "/")
	shortCode := (*shortURL)[slashIndex+1:]
	suite.Len(shortCode, customShortCodeLength)
}

func (suite *CreateServiceSuite) TestSuccess() {
	suite.MockCreateSuccess().Times(1)

	originalURL := "https://www.foo.com"
	shortURL, err := suite.service.CreateShortURL(suite.ctx, originalURL, nil)
	suite.Nil(err)
	suite.NotNil(shortURL)

	// Verify the short URL contains the expected hostname
	suite.Contains(*shortURL, "http://localhost:8080")

	// Verify the short URL contains a valid short code
	slashIndex := strings.LastIndex(*shortURL, "/")
	shortCode := (*shortURL)[slashIndex+1:]
	suite.Len(shortCode, 6)
	suite.NotEmpty(shortCode)
}

func (suite *CreateServiceSuite) MockCreateFail() *gomock.Call {
	return suite.urlRecordDAO.
		EXPECT().
		Create(gomock.Any(), gomock.Any()).
		AnyTimes().
		Return(nil, apperrors.ErrShortCodeAlreadyInUse)
}

func (suite *CreateServiceSuite) MockCreateSuccess() *gomock.Call {
	return suite.urlRecordDAO.
		EXPECT().
		Create(gomock.Any(), gomock.Any()).
		Return(
			&model.URLRecordEntity{
				Entity:    model.Entity{},
				URLRecord: model.URLRecord{},
			},
			nil,
		)
}
