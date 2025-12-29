package read

import (
	"context"
	"errors"
	"testing"
	"tiny-bitly/internal/apperrors"
	"tiny-bitly/internal/config"
	"tiny-bitly/internal/dao"
	mock_daotypes "tiny-bitly/internal/dao/generated"
	"tiny-bitly/internal/model"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

var maxAliasLengthForTest int = 20

type ReadServiceSuite struct {
	suite.Suite
	ctrl         *gomock.Controller
	service      *Service
	dao          dao.DAO
	urlRecordDAO *mock_daotypes.MockURLRecordDAO
}

func TestReadServiceSuite(t *testing.T) {
	suite.Run(t, new(ReadServiceSuite))
}

func (suite *ReadServiceSuite) SetupTest() {
	suite.ctrl = gomock.NewController(suite.T())
	suite.urlRecordDAO = mock_daotypes.NewMockURLRecordDAO(suite.ctrl)
	suite.dao = dao.DAO{
		URLRecordDAO: suite.urlRecordDAO,
	}
	cfg := config.GetTestConfig(config.Config{MaxAliasLength: maxAliasLengthForTest})
	suite.service = NewService(suite.dao, &cfg)
}

func (suite *ReadServiceSuite) TestShortCodeEmpty() {
	shortCode := ""
	originalURL, err := suite.service.GetOriginalURL(context.Background(), shortCode)
	suite.ErrorIs(err, apperrors.ErrShortCodeNotFound)
	suite.Nil(originalURL)
}

func (suite *ReadServiceSuite) TestShortCodeTooLong() {
	shortCode := "0123456789001234567890" // 1 longer than maxAliasLengthForTest
	originalURL, err := suite.service.GetOriginalURL(context.Background(), shortCode)
	suite.ErrorIs(err, apperrors.ErrShortCodeNotFound)
	suite.Nil(originalURL)
}

func (suite *ReadServiceSuite) TestGetByShortCodeError() {
	shortCode := "abc123"
	suite.MockGetError(shortCode, "database error")

	originalURL, err := suite.service.GetOriginalURL(context.Background(), shortCode)
	suite.NotNil(err)
	suite.Nil(originalURL)
	suite.ErrorIs(err, apperrors.ErrDataStoreUnavailable)
}

func (suite *ReadServiceSuite) TestGetByShortCodeNotFound() {
	shortCode := "nonexistent"
	suite.MockGetNotFound(shortCode)

	originalURL, err := suite.service.GetOriginalURL(context.Background(), shortCode)
	suite.NotNil(err)
	suite.Nil(originalURL)
	suite.ErrorIs(err, apperrors.ErrShortCodeNotFound)
}

func (suite *ReadServiceSuite) TestSuccess() {
	shortCode := "abc123"
	expectedOriginalURL := "https://www.example.com"
	suite.MockGetSuccess(shortCode, expectedOriginalURL)

	originalURL, err := suite.service.GetOriginalURL(context.Background(), shortCode)
	suite.Nil(err)
	suite.NotNil(originalURL)
	suite.Equal(expectedOriginalURL, *originalURL)
}

func (suite *ReadServiceSuite) MockGetSuccess(shortCode, expectedOriginalURL string) *gomock.Call {
	return suite.urlRecordDAO.
		EXPECT().
		GetByShortCode(gomock.Any(), shortCode).
		Return(
			&model.URLRecordEntity{
				Entity:    model.Entity{},
				URLRecord: model.URLRecord{OriginalURL: expectedOriginalURL},
			},
			nil,
		)
}

func (suite *ReadServiceSuite) MockGetNotFound(shortCode string) *gomock.Call {
	return suite.urlRecordDAO.
		EXPECT().
		GetByShortCode(gomock.Any(), shortCode).
		Return(nil, nil)
}

func (suite *ReadServiceSuite) MockGetError(shortCode, errorMessage string) *gomock.Call {
	return suite.urlRecordDAO.
		EXPECT().
		GetByShortCode(gomock.Any(), shortCode).
		Return(nil, errors.New(errorMessage))
}
