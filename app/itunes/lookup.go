package itunes

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

func (s *Store) Lookup(id string) (*PodcastDetail, error) {
	resp, err := s.hc.Get(fmt.Sprintf("%s/lookup?id=%s", s.url, id))
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		bytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return nil, errors.New(fmt.Sprintf("itunes api error: %s", string(bytes)))
	}

	r := lookupResponse{}
	if err = json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, err
	}

	if len(r.Results) != 1 {
		return nil, errors.New(fmt.Sprintf("invalid lookup result length: %d, expected 1", len(r.Results)))
	}

	return &PodcastDetail{
		Id:           strconv.Itoa(r.Results[0].Id),
		Artist:       r.Results[0].Artist,
		Name:         r.Results[0].Name,
		Image:        r.Results[0].Image,
		EpisodeCount: r.Results[0].EpisodeCount,
		Url:          r.Results[0].Url,
		FeedUrl:      r.Results[0].FeedUrl,
		Genres:       r.Results[0].Genres,
	}, nil
}

type lookupResponse struct {
	Results []lookupResult `json:"results"`
}

type lookupResult struct {
	Id           int      `json:"collectionId"`
	Artist       string   `json:"artistName"`
	Name         string   `json:"collectionName"`
	Image        string   `json:"artworkUrl600"`
	EpisodeCount int      `json:"trackCount"`
	Url          string   `json:"collectionViewUrl"`
	FeedUrl      string   `json:"feedUrl"`
	Genres       []string `json:"genres"`
}
