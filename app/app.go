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
	mux.HandleFunc("/search", a.handleSearch())
	mux.HandleFunc("/podcasts/", a.handlePodcasts())
	mux.HandleFunc("/", a.handleHome())

	a.mux = mux

	return a
}

func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.mux.ServeHTTP(w, r)
}

func (a *App) handleHome() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			podcasts, err := a.store.Top(region(r))
			if err != nil {
				log.Println(err)
				render(w, r, nil, "./templates/error.html")
				return
			}

			render(w, r, podcasts, "./templates/home.html")
			return
		}

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

func (a *App) handlePodcasts() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		type response struct {
			Podcast *itunes.PodcastDetail
			Reviews []*itunes.Review
		}
		var (
			wg      sync.WaitGroup
			pod     *itunes.PodcastDetail
			podErr  error
			rews    []*itunes.Review
			rewsErr error
		)

		id := ""
		if id = strings.TrimPrefix(r.URL.Path, "/podcasts/"); id == "" {
			render(w, r, nil, "./templates/404.html")
			return
		}

		wg.Add(2)
		go func() {
			defer wg.Done()
			pod, podErr = a.store.Lookup(id)
		}()
		go func() {
			defer wg.Done()
			rews, rewsErr = a.store.Reviews(id, region(r))
		}()
		wg.Wait()

		if podErr != nil {
			log.Println(podErr)
			render(w, r, nil, "./templates/404.html")
			return
		}

		if rewsErr != nil {
			log.Println(rewsErr)
			rews = []*itunes.Review{}
		}

		render(w, r, response{pod, rews}, "./templates/podcast.html")
	}
}

func render(w http.ResponseWriter, r *http.Request, data any, tmpl string) {
	type response struct {
		Data    any
		Region  string
		Regions []itunes.Region
	}
	var (
		init sync.Once
		tpl  *template.Template
		err  error
	)

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
