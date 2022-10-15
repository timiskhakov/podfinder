package itunes

import "time"

const (
	defaultUrl    = "https://itunes.apple.com"
	DefaultRegion = "us"
)

// Regions lists ISO country codes, see: https://en.wikipedia.org/wiki/ISO_3166-1_alpha-2
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
		Value: "gb",
		Name:  "United Kingdom",
	},
	{
		Value: "us",
		Name:  "United States",
	},
}

type Store struct {
	url string
	hc  HttpClient
}

func NewStore(url string, g HttpClient) *Store {
	if url == "" {
		url = defaultUrl
	}

	return &Store{url, g}
}

func isSupportedRegion(v string) bool {
	for _, r := range Regions {
		if r.Value == v {
			return true
		}
	}
	return false
}

type Podcast struct {
	Id     string
	Artist string
	Name   string
	Image  string
}

type PodcastDetail struct {
	Id           string
	Artist       string
	Name         string
	Image        string
	EpisodeCount int
	Url          string
	FeedUrl      string
	Genres       []string
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
