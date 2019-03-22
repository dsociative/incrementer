package api

import (
	context "context"
	"errors"
	"testing"

	"github.com/dsociative/incrementer/db"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
)

type APITestSuite struct {
	suite.Suite

	ctrl *gomock.Controller
	db   *db.MockDB
	api  *API
}

func (s *APITestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.db = db.NewMockDB(s.ctrl)
	s.api = NewAPI(s.db)
}

func (s *APITestSuite) TeardownTest() {
	s.ctrl.Finish()
}

func (s *APITestSuite) TestGetNumber() {
	s.db.EXPECT().Number().Return(555, nil)
	n, err := s.api.GetNumber(context.Background(), nil)
	s.NoError(err)
	s.Equal(int64(555), n.Number)

}

func (s *APITestSuite) TestGetNumberErr() {
	expectedError := errors.New("some error")
	s.db.EXPECT().Number().Return(0, expectedError)
	_, err := s.api.GetNumber(context.Background(), nil)
	s.Equal(expectedError, err)
}

func (s *APITestSuite) TestIncrementNumber() {
	for _, n := range []int{1, 55, 200} {
		s.db.EXPECT().Incr().Return(n, nil)
		number, err := s.api.IncrementNumber(context.Background(), nil)
		s.NoError(err)
		s.Equal(int64(n), number.Number)
	}
}

func (s *APITestSuite) TestIncrementNumberErr() {
	expectedError := errors.New("some error")
	s.db.EXPECT().Incr().Return(0, expectedError)
	_, err := s.api.IncrementNumber(context.Background(), nil)
	s.Equal(expectedError, err)
}

func (s *APITestSuite) TestSetSettings() {
	s.db.EXPECT().SetSettings(10, 1)
	_, err := s.api.SetSettings(context.Background(), &Setting{10, 1})
	s.NoError(err)

	for _, maximum := range []int64{0, -1, -2} {
		_, err = s.api.SetSettings(context.Background(), &Setting{maximum, 1})
		s.Equal(errMaximumBelowZero, err)
	}

	for _, step := range []int64{0, -1, -2} {
		_, err = s.api.SetSettings(context.Background(), &Setting{1, step})
		s.Equal(errStepLessOrEqualZero, err)
	}
}

func TestAPITestSuite(t *testing.T) {
	suite.Run(t, new(APITestSuite))
}
