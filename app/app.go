package main

import (
	"github.com/timiskhakov/podfinder/app/itunes"
	"html/template"
	"log"
	"net/http"
	"strings"
	"sync"
)

const errorMessage = "Internal server error"

type App struct {
	mux   http.Handler
	store Store
}

type Store interface {
	Top(region string) ([]*itunes.Podcast, error)
	Search(region, query string) ([]*itunes.Podcast, error)
	Lookup(id string) (*itunes.PodcastDetail, error)
	Reviews(id, region string) ([]*itunes.Review, error)
}

func NewApp(store Store) *App {
	a := &App{store: store}

	mux := http.NewServeMux()
	mux.Handle("/www/", http.StripPrefix("/www/", http.FileServer(http.Dir("./www/"))))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			a.handleHomePost(w, r)
			return
		}

		if id := strings.TrimPrefix(r.URL.Path, "/"); id != "" {
			a.handlePodcast(w, r, id)
			return
		}

		a.handleHomeGet(w, r)
	})
	mux.HandleFunc("/search", a.handleSearch())

	a.mux = mux

	return a
}

func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.mux.ServeHTTP(w, r)
}

func (a *App) handleHomeGet(w http.ResponseWriter, r *http.Request) {
	podcasts, err := a.store.Top(region(r))
	if err != nil {
		log.Println(err)
		render(w, r, nil, "./templates/error.html")
		return
	}

	render(w, r, podcasts, "./templates/home.html")
}

func (a *App) handleHomePost(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Println(err)
		render(w, r, nil, "./templates/error.html")
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
	type response struct {
		Query    string
		Podcasts []*itunes.Podcast
	}

	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			log.Println(err)
			render(w, r, nil, "./templates/error.html")
			return
		}

		query := r.Form.Get("query")
		podcasts, err := a.store.Search(region(r), query)
		if err != nil {
			log.Println(err)
			render(w, r, nil, "./templates/error.html")
			return
		}

		render(w, r, response{query, podcasts}, "./templates/results.html")
	}
}

func (a *App) handlePodcast(w http.ResponseWriter, r *http.Request, id string) {
	type response struct {
		Podcast *itunes.PodcastDetail
		Reviews []*itunes.Review
	}

	// TODO(timiskhakov): Obtain podcast and reviews at the same time
	podcast, err := a.store.Lookup(id)
	if err != nil {
		log.Println(err)
		render(w, r, nil, "./templates/404.html")
		return
	}

	reviews, err := a.store.Reviews(id, region(r))
	if err != nil {
		log.Println(err)
		reviews = []*itunes.Review{}
	}

	render(w, r, response{podcast, reviews}, "./templates/podcast.html")
}

func render(w http.ResponseWriter, r *http.Request, data any, tmpl string) {
	var (
		init sync.Once
		tpl  *template.Template
		err  error
	)
	type response struct {
		Data    any
		Region  string
		Regions []itunes.Region
	}

	init.Do(func() {
		tpl, err = template.ParseFiles("./templates/base.html", tmpl)
	})
	if err != nil {
		log.Println(err)
		http.Error(w, errorMessage, http.StatusInternalServerError)
		return
	}

	if err = tpl.Execute(w, &response{
		Data:    data,
		Region:  region(r),
		Regions: itunes.Regions,
	}); err != nil {
		log.Println(err)
		http.Error(w, errorMessage, http.StatusInternalServerError)
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
