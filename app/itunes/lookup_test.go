package itunes

import (
	"github.com/golang/mock/gomock"
	"github.com/timiskhakov/podfinder/app/itunes/mock"
	"net/http"
	"os"
)

func (s *StoreSuite) TestLookup() {
	fh, err := os.Open("../testdata/lookup.json")
	s.NoError(err)
	defer func() { _ = fh.Close() }()
	g := mock.NewMockHttpClient(s.ctrl)
	g.EXPECT().Get(gomock.Any()).Return(&http.Response{
		StatusCode: http.StatusOK,
		Body:       fh,
	}, nil)
	store := Store{"", g}

	pd, err := store.Lookup("811377230")

	s.NoError(err)
	s.Equal(&PodcastDetail{
		Id:           "811377230",
		Artist:       "CGP Grey & Brady Haran",
		Name:         "Hello Internet",
		Image:        "https://is5-ssl.mzstatic.com/image/thumb/Podcasts6/v4/19/33/fe/1933fe85-cd86-2191-8187-d725ca7359bf/mza_8038397602264410223.png/600x600bb.jpg",
		EpisodeCount: 100,
		Url:          "https://podcasts.apple.com/us/podcast/hello-internet/id811377230?uo=4",
		FeedUrl:      "http://www.hellointernet.fm/podcast?format=rss",
		Genres:       []string{"Education", "Podcasts"},
	}, pd)
}
