package main

import (
	"github.com/timiskhakov/podfinder/app/itunes"
	"html/template"
	"log"
	"net/http"
	"strings"
)

type App struct {
	router http.Handler
	store  Store
}

type Store interface {
	Top(region string) ([]*itunes.Podcast, error)
	Search(region, query string) ([]*itunes.Podcast, error)
	Lookup(id string) (*itunes.PodcastDetail, error)
	Reviews(id, region string) ([]*itunes.Review, error)
}

func NewApp(store Store) *App {
	a := &App{store: store}

	r := http.NewServeMux()
	r.Handle("/www/", http.StripPrefix("/www/", http.FileServer(http.Dir("./www/"))))

	r.HandleFunc("/", a.handleHome())
	r.HandleFunc("/search", a.handleSearch())

	a.router = r

	return a
}

func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.router.ServeHTTP(w, r)
}

func (a *App) handleHome() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			a.handleHomePost(w, r)
			return
		}

		if id := strings.TrimPrefix(r.URL.Path, "/"); id != "" {
			a.handlePodcast(w, r, id)
			return
		}

		a.handleHomeGet(w, r)
	}
}

func (a *App) handleHomeGet(w http.ResponseWriter, r *http.Request) {
	podcasts, err := a.store.Top(region(r))
	if err != nil {
		render(w, r, nil, err, "./templates/base.html", "./templates/error.html")
		return
	}

	render(w, r, podcasts, nil, "./templates/base.html", "./templates/home.html")
}

func (a *App) handleHomePost(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		render(w, r, nil, err, "./templates/base.html", "./templates/error.html")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "region",
		Value:    r.Form.Get("region"),
		HttpOnly: true,
	})

	http.Redirect(w, r, r.RequestURI, 301)
}

func (a *App) handleSearch() http.HandlerFunc {
	type queryAndPodcasts struct {
		Query    string
		Podcasts []*itunes.Podcast
	}

	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			render(w, r, nil, err, "./templates/base.html", "./templates/error.html")
			return
		}

		query := r.Form.Get("query")
		podcasts, err := a.store.Search(region(r), query)
		if err != nil {
			render(w, r, nil, err, "./templates/base.html", "./templates/error.html")
			return
		}

		render(w, r, queryAndPodcasts{query, podcasts}, nil, "./templates/base.html", "./templates/results.html")
	}
}

func (a *App) handlePodcast(w http.ResponseWriter, r *http.Request, id string) {
	type podcastAndReviews struct {
		Podcast *itunes.PodcastDetail
		Reviews []*itunes.Review
	}

	// TODO(timiskhakov): Obtain podcast and reviews at the same time
	podcast, err := a.store.Lookup(id)
	if err != nil {
		render(w, r, nil, nil, "./templates/base.html", "./templates/404.html")
		return
	}

	reviews, err := a.store.Reviews(id, region(r))
	if err != nil {
		log.Println(err.Error())
		reviews = []*itunes.Review{}
	}

	render(w, r, podcastAndReviews{podcast, reviews}, nil, "./templates/base.html", "./templates/podcast.html")

}

func render(w http.ResponseWriter, r *http.Request, data any, err error, templates ...string) {
	if err != nil {
		log.Println(err.Error())
	}

	ts, err := template.ParseFiles(templates...)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "Internal Server Error", 500)
		return
	}

	if err = ts.Execute(w, &response{
		Region:  region(r),
		Regions: itunes.Regions,
		Data:    data,
	}); err != nil {
		log.Println(err.Error())
		http.Error(w, "Internal Server Error", 500)
		return
	}
}

func region(r *http.Request) string {
	cookie, err := r.Cookie("region")
	if err != nil || cookie.Value == "" {
		return itunes.DefaultRegion
	}

	return cookie.Value
}

type response struct {
	Region  string
	Regions []itunes.Region
	Data    any
}
