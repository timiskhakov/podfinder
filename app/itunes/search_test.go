package itunes

import (
	"github.com/golang/mock/gomock"
	"github.com/timiskhakov/podfinder/app/itunes/mock"
	"net/http"
	"os"
)

func (s *StoreSuite) TestSearch() {
	fh, err := os.Open("../testdata/search.json")
	s.NoError(err)
	defer fh.Close()
	g := mock.NewMockHttpClient(s.ctrl)
	g.EXPECT().Get(gomock.Any()).Return(&http.Response{
		StatusCode: http.StatusOK,
		Body:       fh,
	}, nil)
	store := Store{"", g}

	podcasts, err := store.Search("", "Hello Internet")

	s.NoError(err)
	s.Equal(5, len(podcasts))
	s.Equal(&Podcast{
		Id:     "811377230",
		Artist: "CGP Grey & Brady Haran",
		Name:   "Hello Internet",
		Image:  "https://is5-ssl.mzstatic.com/image/thumb/Podcasts6/v4/19/33/fe/1933fe85-cd86-2191-8187-d725ca7359bf/mza_8038397602264410223.png/600x600bb.jpg",
	}, podcasts[0])
}
