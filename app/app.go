package main

import (
	"fmt"
	"github.com/timiskhakov/podfinder/app/itunes"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
)

const errorMessage = "Internal server error"

type App struct {
	store            Store
	isLimiterEnabled bool
	limiter          Limiter
	mux              http.Handler
	cache            map[string]*template.Template
}

type Store interface {
	Top(region string) ([]*itunes.Podcast, error)
	Search(region, query string) ([]*itunes.Podcast, error)
	Lookup(id string) (*itunes.PodcastDetail, error)
	Reviews(id, region string) ([]*itunes.Review, error)
}

type Limiter interface {
	Allow() bool
}

type AppConfig struct {
	Store            Store
	IsLimiterEnabled bool
	Limiter          Limiter
}

func NewApp(config *AppConfig) (*App, error) {
	a := &App{
		store:            config.Store,
		isLimiterEnabled: config.IsLimiterEnabled,
		limiter:          config.Limiter,
	}

	mux := http.NewServeMux()
	mux.Handle("/www/", http.StripPrefix("/www/", http.FileServer(http.Dir("./www/"))))
	mux.HandleFunc("/search", a.limit(a.handleSearch()))
	mux.HandleFunc("/podcasts/", a.handlePodcasts())
	mux.HandleFunc("/", a.handleHome())
	a.mux = mux

	pages, err := filepath.Glob("./templates/*.html")
	if err != nil {
		return nil, err
	}

	cache := make(map[string]*template.Template, len(pages))
	for _, page := range pages {
		ts, err := template.ParseFiles("./templates/base.html", page)
		if err != nil {
			return nil, err
		}

		cache[filepath.Base(page)] = ts
	}

	a.cache = cache

	return a, nil
}

func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.mux.ServeHTTP(w, r)
}

func (a *App) handleHome() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			podcasts, err := a.store.Top(region(r))
			if err != nil {
				log.Printf("%v", err)
				a.render(w, r, nil, "error.html")
				return
			}

			a.render(w, r, podcasts, "home.html")
			return
		}

		if err := r.ParseForm(); err != nil {
			log.Printf("%v", err)
			a.render(w, r, nil, "error.html")
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "region",
			Value:    r.Form.Get("region"),
			HttpOnly: true,
		})
		http.Redirect(w, r, r.Referer(), 301)
	}
}

func (a *App) handleSearch() http.HandlerFunc {
	type response struct {
		Query    string
		Podcasts []*itunes.Podcast
	}

	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			log.Printf("%v", err)
			a.render(w, r, nil, "error.html")
			return
		}

		query := r.Form.Get("query")
		podcasts, err := a.store.Search(region(r), query)
		if err != nil {
			log.Printf("%v", err)
			a.render(w, r, nil, "error.html")
			return
		}

		a.render(w, r, response{query, podcasts}, "results.html")
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
			a.render(w, r, nil, "404.html")
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
			a.render(w, r, nil, "404.html")
			return
		}

		if rewsErr != nil {
			log.Println(rewsErr)
			rews = []*itunes.Review{}
		}

		a.render(w, r, response{pod, rews}, "podcast.html")
	}
}

func (a *App) limit(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if a.isLimiterEnabled && !a.limiter.Allow() {
			a.render(w, r, nil, "limit.html")
			return
		}

		next(w, r)
	}
}

func (a *App) render(w http.ResponseWriter, r *http.Request, data any, tmpl string) {
	type response struct {
		Data    any
		Region  string
		Regions []itunes.Region
	}

	t, ok := a.cache[tmpl]
	if !ok {
		log.Printf(fmt.Sprintf("can't find template %s", tmpl))
		http.Error(w, errorMessage, http.StatusInternalServerError)
		return
	}

	if err := t.Execute(w, &response{
		Data:    data,
		Region:  region(r),
		Regions: itunes.Regions,
	}); err != nil {
		log.Printf("%v", err)
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
