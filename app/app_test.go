package main

import (
	"fmt"
	"github.com/stretchr/testify/suite"
	"github.com/timiskhakov/podfinder/app/itunes"
	"golang.org/x/time/rate"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"
)

type AppSuite struct {
	suite.Suite
	itunesMux    *http.ServeMux
	itunesServer *httptest.Server
	httpClient   *http.Client
	app          *App
	appServer    *httptest.Server
}

func (s *AppSuite) SetupTest() {
	s.itunesMux = http.NewServeMux()
	s.itunesServer = httptest.NewServer(s.itunesMux)
	s.httpClient = s.itunesServer.Client()

	app, err := NewApp(&AppConfig{
		Store:            itunes.NewStore(s.itunesServer.URL, s.httpClient),
		IsLimiterEnabled: false,
		Limiter:          &rate.Limiter{},
		InfoLog:          log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime),
		ErrorLog:         log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile),
	})
	s.NoError(err)

	s.app = app
	s.appServer = httptest.NewServer(app)
}

func (s *AppSuite) TearDownTest() {
	s.appServer.Close()
	s.itunesServer.Close()
}

func TestAppSuite(t *testing.T) {
	suite.Run(t, new(AppSuite))
}

func (s *AppSuite) TestNewApp() {
	s.NotNil(s.app)
	s.Equal(7, len(s.app.cache))
}

func (s *AppSuite) TestHandleHomeGet() {
	s.itunesMux.HandleFunc("/us/rss/toppodcasts/limit=10/json", func(w http.ResponseWriter, r *http.Request) {
		f, err := os.Open("./testdata/lookup.json")
		s.NoError(err)
		defer func() { _ = f.Close() }()
		bytes, err := io.ReadAll(f)
		s.NoError(err)
		_, _ = w.Write(bytes)
	})

	resp, err := s.httpClient.Get(s.appServer.URL)
	s.NoError(err)
	s.Equal(http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	s.NoError(err)
	s.Contains(string(body), "Top podcasts")
}

func (s *AppSuite) TestHandleHomePost() {
	s.httpClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	resp, err := s.httpClient.PostForm(s.appServer.URL, url.Values{"region": []string{"fi"}})
	s.NoError(err)
	s.Equal(http.StatusMovedPermanently, resp.StatusCode)
}

func (s *AppSuite) TestHandleSearch() {
	s.itunesMux.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		f, err := os.Open("./testdata/search.json")
		s.NoError(err)
		defer func() { _ = f.Close() }()

		bytes, err := io.ReadAll(f)
		s.NoError(err)

		_, err = w.Write(bytes)
		s.NoError(err)
	})

	resp, err := s.httpClient.Get(fmt.Sprintf("%s/search?query=hello+internet", s.appServer.URL))

	s.NoError(err)
	s.Equal(http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	s.NoError(err)
	s.Contains(string(body), "Search results")
}

func (s *AppSuite) TestHandlePodcast() {
	s.itunesMux.HandleFunc("/lookup", func(w http.ResponseWriter, r *http.Request) {
		f, err := os.Open("./testdata/lookup.json")
		s.NoError(err)
		defer func() { _ = f.Close() }()

		bytes, err := io.ReadAll(f)
		s.NoError(err)

		_, err = w.Write(bytes)
		s.NoError(err)
	})
	s.itunesMux.HandleFunc("/us/rss/customerreviews/id=123/json", func(w http.ResponseWriter, r *http.Request) {
		f, err := os.Open("./testdata/reviews.json")
		s.NoError(err)
		defer func() { _ = f.Close() }()

		bytes, err := io.ReadAll(f)
		s.NoError(err)

		_, err = w.Write(bytes)
		s.NoError(err)
	})

	resp, err := s.httpClient.Get(fmt.Sprintf("%s/podcasts/123", s.appServer.URL))

	s.NoError(err)
	s.Equal(http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	s.NoError(err)
	s.Contains(string(body), "Hello Internet")            // Podcast info is in the page
	s.Contains(string(body), "Re-listening to the show.") // Review is in the page
}

func (s *AppSuite) TestLimit() {
	s.app.isLimiterEnabled = true
	s.app.limiter = rate.NewLimiter(rate.Every(1*time.Minute), 0)

	resp, err := s.httpClient.Get(fmt.Sprintf("%s/search?query=hello+internet", s.appServer.URL))
	s.NoError(err)
	s.Equal(http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	s.NoError(err)
	s.Contains(string(body), "An iTunes request limit has been reached")
}
