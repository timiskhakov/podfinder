package itunes

import (
	"github.com/golang/mock/gomock"
	"github.com/timiskhakov/podfinder/app/itunes/mock"
	"net/http"
	"os"
)

func (s *StoreSuite) TestTop() {
	fh, err := os.Open("../testdata/top.json")
	s.NoError(err)
	defer fh.Close()
	g := mock.NewMockGetter(s.ctrl)
	g.EXPECT().Get(gomock.Any()).Return(&http.Response{
		StatusCode: http.StatusOK,
		Body:       fh,
	}, nil)
	store := Store{"", g}

	podcasts, err := store.Top("")

	s.NoError(err)
	s.Equal(10, len(podcasts))
	s.Equal(&Podcast{
		Id:     "1612875889",
		Artist: "HLN",
		Name:   "Very Scary People",
		Image:  "https://is2-ssl.mzstatic.com/image/thumb/Podcasts116/v4/18/69/79/18697926-b149-c6e0-d33c-ce6fb250efec/mza_17914905586066761253.jpg/170x170bb.png",
	}, podcasts[0])
}

func (s *StoreSuite) TestSelectBiggestImage() {
	images := []image{
		{
			Attributes: heightAttributes{
				Height: "170",
			},
			Label: "170x170.png",
		},
		{
			Attributes: heightAttributes{
				Height: "250",
			},
			Label: "250x250.png",
		},
		{
			Attributes: heightAttributes{
				Height: "500",
			},
			Label: "500x500.png",
		},
	}

	s.Equal("500x500.png", selectBiggestImage(images))
}
