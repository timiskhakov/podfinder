package main

import (
	"fmt"
	"github.com/stretchr/testify/suite"
	"github.com/timiskhakov/podfinder/app/itunes"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
)

const (
	rgn       = "us"
	podcastId = "123"
)

type AppSuite struct {
	suite.Suite
	mux  *http.ServeMux
	itns *httptest.Server
	hc   *http.Client
	srv  *httptest.Server
}

func (s *AppSuite) SetupTest() {
	s.mux = http.NewServeMux()
	s.itns = httptest.NewServer(s.mux)
	s.hc = s.itns.Client()

	app, err := NewApp(itunes.NewStore(s.itns.URL, s.hc))
	s.NoError(err)

	s.srv = httptest.NewServer(app)
}

func (s *AppSuite) TearDownTest() {
	s.srv.Close()
	s.itns.Close()
}

func TestAppSuite(t *testing.T) {
	suite.Run(t, new(AppSuite))
}

func (s *AppSuite) TestNewApp() {
	s.NotNil(s.srv)
}

func (s *AppSuite) TestHandleHome() {
	s.mux.HandleFunc(fmt.Sprintf("/%s/rss/toppodcasts/limit=10/json", rgn), func(w http.ResponseWriter, r *http.Request) {
		f, err := os.Open("./testdata/lookup.json")
		s.NoError(err)
		defer func() { _ = f.Close() }()
		bytes, err := io.ReadAll(f)
		s.NoError(err)
		_, _ = w.Write(bytes)
	})

	resp, err := s.hc.Get(s.srv.URL)

	s.NoError(err)
	s.Equal(http.StatusOK, resp.StatusCode)
}

func (s *AppSuite) TestHandleRegion() {
	s.hc.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	resp, err := s.hc.PostForm(s.srv.URL, url.Values{"region": []string{"fi"}})

	s.NoError(err)
	s.Equal(http.StatusMovedPermanently, resp.StatusCode)
}

func (s *AppSuite) TestHandleSearch() {
	s.mux.HandleFunc(fmt.Sprintf("/search?media=podcast&entity=podcast&country=%s&term=%s", rgn, "Hello+Internet"), func(w http.ResponseWriter, r *http.Request) {
		f, err := os.Open("./testdata/search.json")
		s.NoError(err)
		defer func() { _ = f.Close() }()
		bytes, err := io.ReadAll(f)
		s.NoError(err)
		_, _ = w.Write(bytes)
	})

	resp, err := s.hc.Get(fmt.Sprintf("%s/search", s.srv.URL))

	s.NoError(err)
	s.Equal(http.StatusOK, resp.StatusCode)
}

func (s *AppSuite) TestHandlePodcast() {
	s.mux.HandleFunc(fmt.Sprintf("/lookup?id=%s", podcastId), func(w http.ResponseWriter, r *http.Request) {
		f, err := os.Open("./testdata/lookup.json")
		s.NoError(err)
		defer func() { _ = f.Close() }()
		bytes, err := io.ReadAll(f)
		s.NoError(err)
		_, _ = w.Write(bytes)
	})
	s.mux.HandleFunc(fmt.Sprintf("/%s/rss/customerreviews/id=%s/json", rgn, podcastId), func(w http.ResponseWriter, r *http.Request) {
		f, err := os.Open("./testdata/reviews.json")
		s.NoError(err)
		defer func() { _ = f.Close() }()
		bytes, err := io.ReadAll(f)
		s.NoError(err)
		_, _ = w.Write(bytes)
	})

	resp, err := s.hc.Get(fmt.Sprintf("%s/123", s.srv.URL))

	s.NoError(err)
	s.Equal(http.StatusOK, resp.StatusCode)
}
