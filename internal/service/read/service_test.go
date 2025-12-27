package read

import (
	"context"
	"errors"
	"testing"
	"tiny-bitly/internal/dao"
	mock_daotypes "tiny-bitly/internal/dao/generated"
	"tiny-bitly/internal/model"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

type ReadServiceSuite struct {
	suite.Suite
	ctrl         *gomock.Controller
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
}

func (suite *ReadServiceSuite) TestEmptyShortCode() {
	shortCode := ""
	originalURL, err := GetOriginalURL(context.Background(), suite.dao, shortCode)
	suite.Nil(err)
	suite.Nil(originalURL)
}

func (suite *ReadServiceSuite) TestGetByShortCodeError() {
	shortCode := "abc123"
	suite.urlRecordDAO.
		EXPECT().
		GetByShortCode(gomock.Any(), shortCode).
		Return(nil, errors.New("database error"))

	originalURL, err := GetOriginalURL(context.Background(), suite.dao, shortCode)
	suite.NotNil(err)
	suite.Nil(originalURL)
	suite.ErrorContains(err, "failed to get original URL by short code")
}

func (suite *ReadServiceSuite) TestGetByShortCodeNotFound() {
	shortCode := "nonexistent"
	suite.urlRecordDAO.
		EXPECT().
		GetByShortCode(gomock.Any(), shortCode).
		Return(nil, nil)

	originalURL, err := GetOriginalURL(context.Background(), suite.dao, shortCode)
	suite.Nil(err)
	suite.Nil(originalURL)
}

func (suite *ReadServiceSuite) TestSuccess() {
	shortCode := "abc123"
	expectedOriginalURL := "https://www.example.com"
	suite.urlRecordDAO.
		EXPECT().
		GetByShortCode(gomock.Any(), shortCode).
		Return(
			&model.URLRecordEntity{
				Entity:    model.Entity{},
				URLRecord: model.URLRecord{OriginalURL: expectedOriginalURL},
			},
			nil,
		)

	originalURL, err := GetOriginalURL(context.Background(), suite.dao, shortCode)
	suite.Nil(err)
	suite.NotNil(originalURL)
	suite.Equal(expectedOriginalURL, *originalURL)
}
