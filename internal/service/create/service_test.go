package create

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"tiny-bitly/internal/apperrors"
	"tiny-bitly/internal/dao"
	mock_daotypes "tiny-bitly/internal/dao/generated"
	"tiny-bitly/internal/model"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

type CreateServiceSuite struct {
	suite.Suite
	ctrl         *gomock.Controller
	dao          dao.DAO
	urlRecordDAO *mock_daotypes.MockURLRecordDAO
}

func TestCreateServiceSuite(t *testing.T) {
	suite.Run(t, new(CreateServiceSuite))
}

func (suite *CreateServiceSuite) SetupTest() {
	os.Clearenv()
	os.Setenv("API_HOSTNAME", "http://localhost:8080")
	suite.ctrl = gomock.NewController(suite.T())
	suite.urlRecordDAO = mock_daotypes.NewMockURLRecordDAO(suite.ctrl)
	suite.dao = dao.DAO{
		URLRecordDAO: suite.urlRecordDAO,
	}
}

func (suite *CreateServiceSuite) TestErrorInputURLEmpty() {
	originalURL := ""
	_, err := createShortURL(context.Background(), suite.dao, originalURL, nil)
	suite.NotNil(err)
	suite.ErrorContains(err, "invalid URL")
}

func (suite *CreateServiceSuite) TestErrorInputURLInvalidChars() {
	originalURL := "www.`.com"
	_, err := createShortURL(context.Background(), suite.dao, originalURL, nil)
	suite.NotNil(err)
	suite.ErrorContains(err, "invalid URL")
}

func (suite *CreateServiceSuite) TestErrorInputURLTooLong() {
	os.Setenv("MAX_URL_LENGTH", "2")
	originalURL := "abc"
	_, err := createShortURL(context.Background(), suite.dao, originalURL, nil)
	suite.NotNil(err)
	suite.ErrorContains(err, "URL must be shorter than 2 chars")
}

func (suite *CreateServiceSuite) TestErrorInputAliasEmpty() {
	originalURL := "https://www.foo.com"
	alias := ""
	_, err := createShortURL(context.Background(), suite.dao, originalURL, &alias)
	suite.NotNil(err)
	suite.ErrorContains(err, "invalid alias")
}

func (suite *CreateServiceSuite) TestErrorInputAliasInvalidChars() {
	originalURL := "https://www.foo.com"
	alias := "`"
	_, err := createShortURL(context.Background(), suite.dao, originalURL, &alias)
	suite.NotNil(err)
	suite.ErrorContains(err, "invalid alias")
}

func (suite *CreateServiceSuite) TestErrorInputAliasAlreadyUsedForDifferentURL() {
	// Mock: Create() should return a specific error code.
	suite.urlRecordDAO.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, apperrors.ErrShortCodeAlreadyInUse)

	originalURL := "https://www.foo.com"
	alias := "bar"
	_, err := createShortURL(context.Background(), suite.dao, originalURL, &alias)
	suite.NotNil(err)
	suite.ErrorContains(err, "custom alias already in use")
}

func (suite *CreateServiceSuite) TestErrorConfigAPIHostnameMissing() {
	os.Clearenv()

	originalURL := "https://www.foo.com"
	_, err := createShortURL(context.Background(), suite.dao, originalURL, nil)
	suite.NotNil(err)
	suite.ErrorContains(err, "environment variable must be configured: API_HOSTNAME")
}

func (suite *CreateServiceSuite) TestErrorMaxRetries() {
	// Set a low max tries to test the retry limit.
	os.Setenv("MAX_TRIES_CREATE_SHORT_CODE", "3")

	// Mock: Create() should fail to generate an unused short code.
	suite.urlRecordDAO.
		EXPECT().
		Create(gomock.Any(), gomock.Any()).
		AnyTimes().
		Return(nil, apperrors.ErrShortCodeAlreadyInUse)

	originalURL := "https://www.foo.com"
	_, err := createShortURL(context.Background(), suite.dao, originalURL, nil)
	suite.NotNil(err)
	suite.ErrorContains(err, "failed to generate unique short code after maximum retries")
}

func (suite *CreateServiceSuite) TestConfigMaxTries() {
	// Allow only a single attempt.
	os.Setenv("MAX_TRIES_CREATE_SHORT_CODE", "1")

	// First call fails.
	suite.urlRecordDAO.
		EXPECT().
		Create(gomock.Any(), gomock.Any()).
		Return(nil, apperrors.ErrShortCodeAlreadyInUse).
		Times(1)

	// Second call must NEVER happen.
	suite.urlRecordDAO.
		EXPECT().
		Create(gomock.Any(), gomock.Any()).
		Return(
			&model.URLRecordEntity{
				Entity:    model.Entity{},
				URLRecord: model.URLRecord{ShortCode: "myUniqueCode"},
			},
			nil,
		).
		Times(0)

	originalURL := "https://www.foo.com"
	_, err := createShortURL(context.Background(), suite.dao, originalURL, nil)

	suite.Error(err)
	suite.ErrorContains(err, "failed to generate unique short code after maximum retries")
}

func (suite *CreateServiceSuite) TestConfigShortCodeLength() {
	customShortCodeLength := 8
	os.Setenv("SHORT_CODE_LENGTH", fmt.Sprintf("%d", customShortCodeLength))

	suite.urlRecordDAO.
		EXPECT().
		Create(gomock.Any(), gomock.Any()).
		Return(
			&model.URLRecordEntity{
				Entity:    model.Entity{},
				URLRecord: model.URLRecord{},
			},
			nil,
		).
		Times(1)

	originalURL := "https://www.foo.com"
	shortURL, err := createShortURL(context.Background(), suite.dao, originalURL, nil)
	suite.Nil(err)
	suite.NotNil(shortURL)
	slashIndex := strings.LastIndex(*shortURL, "/")
	shortCode := (*shortURL)[slashIndex+1:]
	suite.Len(shortCode, customShortCodeLength)
}

func (suite *CreateServiceSuite) TestSuccess() {
	suite.urlRecordDAO.
		EXPECT().
		Create(gomock.Any(), gomock.Any()).
		Return(
			&model.URLRecordEntity{
				Entity:    model.Entity{},
				URLRecord: model.URLRecord{},
			},
			nil,
		).
		Times(1)

	originalURL := "https://www.foo.com"
	shortURL, err := createShortURL(context.Background(), suite.dao, originalURL, nil)
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
