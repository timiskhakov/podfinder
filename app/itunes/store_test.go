package itunes

import (
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
	"github.com/timiskhakov/podfinder/app/itunes/mock"
	"testing"
)

type StoreSuite struct {
	suite.Suite
	ctrl *gomock.Controller
}

func TestClientSuite(t *testing.T) {
	suite.Run(t, new(StoreSuite))
}

func (s *StoreSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
}

func (s *StoreSuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *StoreSuite) TestNewStore() {
	g := mock.NewMockGetter(s.ctrl)

	store := NewStore("", g)

	s.NotNil(store)
	s.Equal(defaultUrl, store.url)
}

func (s *StoreSuite) TestIsSupportedRegion() {
	cases := []struct {
		region      string
		isSupported bool
	}{
		{"us", true},
		{"uk", false},
	}

	for _, c := range cases {
		c := c
		s.Run(c.region, func() {
			s.Equal(c.isSupported, isSupportedRegion(c.region))
		})
	}
}
