package itunes

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

func (s *Store) Search(region, query string) ([]*Podcast, error) {
	if region == "" || !isSupportedRegion(region) {
		region = DefaultRegion
	}

	resp, err := s.hc.Get(fmt.Sprintf("%s/search?media=podcast&entity=podcast&country=%s&term=%s", s.url, region, url.QueryEscape(query)))
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		bytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("itunes search api error: %s", string(bytes))
	}

	r := searchResponse{}
	if err = json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, err
	}

	podcasts := make([]*Podcast, len(r.Results))
	for i, r := range r.Results {
		podcasts[i] = &Podcast{
			Id:     strconv.Itoa(r.Id),
			Artist: r.Artist,
			Name:   r.Name,
			Image:  r.Image,
		}
	}

	return podcasts, nil
}

type searchResponse struct {
	Count   int            `json:"resultCount"`
	Results []searchResult `json:"results"`
}

type searchResult struct {
	Id     int    `json:"collectionId"`
	Artist string `json:"artistName"`
	Name   string `json:"collectionName"`
	Image  string `json:"artworkUrl600"`
}
