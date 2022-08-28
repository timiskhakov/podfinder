package itunes

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

func (s *Store) Reviews(id, region string) ([]*Review, error) {
	if region == "" || !isSupportedRegion(region) {
		region = DefaultRegion
	}

	resp, err := s.hc.Get(fmt.Sprintf("%s/%s/rss/customerreviews/id=%s/json", s.url, region, id))
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		bytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("itunes reviews api error: %s", string(bytes))
	}

	r := reviewsResponse{}
	if err = json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, err
	}

	reviews := make([]*Review, len(r.Feed.Reviews))
	for i, r := range r.Feed.Reviews {
		reviews[i] = &Review{
			Id:      r.Id.Label,
			Author:  r.Author.Name.Label,
			Title:   r.Title.Label,
			Content: r.Content.Label,
			Rating:  rating(r.Rating.Label),
			Date:    date(r.Updated.Label),
		}
	}

	return reviews, nil
}

func rating(s string) []struct{} {
	v, _ := strconv.Atoi(s)
	return make([]struct{}, v)
}

func date(s string) time.Time {
	v, _ := time.Parse(time.RFC3339, s)
	return v
}

type reviewsResponse struct {
	Feed reviewsFeed `json:"feed"`
}

type reviewsFeed struct {
	Update  text     `json:"updated"`
	Reviews []review `json:"entry"`
}

type review struct {
	Id      text   `json:"id"`
	Title   text   `json:"title"`
	Author  author `json:"author"`
	Content text   `json:"content"`
	Rating  text   `json:"im:rating"`
	Updated text   `json:"updated"`
}

type author struct {
	Name text `json:"name"`
}
