package itunes

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

func (s *Store) Top(region string) ([]*Podcast, error) {
	if region == "" || !isSupportedRegion(region) {
		region = DefaultRegion
	}

	resp, err := s.hc.Get(fmt.Sprintf("%s/%s/rss/toppodcasts/limit=10/json", s.url, region))
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		bytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("itunes toppodcasts api error: %s", string(bytes))
	}

	r := topResponse{}
	if err = json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, err
	}

	podcasts := make([]*Podcast, len(r.Feed.Podcasts))
	for i, p := range r.Feed.Podcasts {
		podcasts[i] = &Podcast{
			Id:     p.Id.Attributes.Id,
			Artist: p.Artist.Label,
			Name:   p.Name.Label,
			Image:  selectBiggestImage(p.Images),
		}
	}

	return podcasts, nil
}

func selectBiggestImage(images []image) string {
	if len(images) == 0 {
		return ""
	}

	biggest, _ := strconv.Atoi(images[0].Attributes.Height)
	result := images[0].Label
	for _, image := range images {
		value, _ := strconv.Atoi(image.Attributes.Height)
		if value > biggest {
			result = image.Label
		}
	}

	return result
}

type topResponse struct {
	Feed topFeed `json:"feed"`
}

type topFeed struct {
	Update   text      `json:"updated"`
	Podcasts []podcast `json:"entry"`
}

type podcast struct {
	Id     id      `json:"id"`
	Artist text    `json:"im:artist"`
	Name   text    `json:"im:name"`
	Images []image `json:"im:image"`
}

type text struct {
	Label string `json:"label"`
}

type id struct {
	Attributes idAttributes `json:"attributes"`
}

type idAttributes struct {
	Id string `json:"im:id"`
}

type image struct {
	Attributes heightAttributes `json:"attributes"`
	Label      string           `json:"label"`
}

type heightAttributes struct {
	Height string `json:"height"`
}
