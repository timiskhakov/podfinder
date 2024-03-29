package itunes

import (
	"github.com/golang/mock/gomock"
	"github.com/timiskhakov/podfinder/app/itunes/mock"
	"net/http"
	"os"
)

func (s *StoreSuite) TestReviews() {
	fh, err := os.Open("../testdata/reviews.json")
	s.NoError(err)
	defer func() { _ = fh.Close() }()
	g := mock.NewMockHttpClient(s.ctrl)
	g.EXPECT().Get(gomock.Any()).Return(&http.Response{
		StatusCode: http.StatusOK,
		Body:       fh,
	}, nil)
	store := Store{"", g}

	reviews, err := store.Reviews("811377230", "us")

	s.NoError(err)
	s.Equal(50, len(reviews))
}

func (s *StoreSuite) TestRating() {
	cases := []struct {
		s string
		v int
	}{
		{"5", 5},
		{"invalid", 0},
	}

	for _, c := range cases {
		c := c
		s.Run(c.s, func() {
			s.Equal(c.v, len(rating(c.s)))
		})
	}
}
