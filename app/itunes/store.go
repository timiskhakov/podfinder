package itunes

import "time"

const (
	defaultUrl    = "https://itunes.apple.com"
	defaultRegion = "us"
)

var Regions = []Region{
	{
		Value: "fi",
		Name:  "Finland",
	},
	{
		Value: "ru",
		Name:  "Russia",
	},
	{
		Value: "us",
		Name:  "United States",
	},
}

type Store struct {
	url string
	hc  Getter
}

func NewStore(url string, g Getter) *Store {
	if url == "" {
		url = defaultUrl
	}

	return &Store{url, g}
}

type Podcast struct {
	Id     string
	Artist string
	Name   string
	Image  string
}

type PodcastDetail struct {
	Id      string
	Artist  string
	Name    string
	Image   string
	Url     string
	FeedUrl string
	Genres  []string
}

type Review struct {
	Id      string
	Author  string
	Title   string
	Content string
	Rating  []struct{}
	Date    time.Time
}

type Region struct {
	Value string
	Name  string
}
