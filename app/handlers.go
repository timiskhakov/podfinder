package main

import (
	"github.com/gorilla/mux"
	"github.com/timiskhakov/podfinder/app/itunes"
	"html/template"
	"log"
	"net/http"
)

type app struct {
	router http.Handler
	str    store
}

type store interface {
	Top(region string) ([]*itunes.Podcast, error)
	Search(region, query string) ([]*itunes.Podcast, error)
	Lookup(id string) (*itunes.PodcastDetail, error)
	Reviews(id, region string) ([]*itunes.Review, error)
}

func NewApp(str store) *app {
	a := &app{str: str}

	r := mux.NewRouter()
	r.PathPrefix("/www/").Handler(http.StripPrefix("/www/", http.FileServer(http.Dir("./www/"))))

	r.HandleFunc("/", a.handleHome()).Methods(http.MethodGet)
	r.HandleFunc("/", a.handleRegion()).Methods(http.MethodPost)
	r.HandleFunc("/search", a.handleSearch()).Methods(http.MethodGet)
	r.HandleFunc("/{id:[0-9]+}", a.handlePodcast()).Methods(http.MethodGet)

	a.router = r

	return a
}

func (a *app) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.router.ServeHTTP(w, r)
}

func (a *app) handleHome() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rgn := region(r)
		podcasts, err := a.str.Top(rgn)
		if err != nil {
			log.Println(err.Error())
			http.Error(w, "Internal Server Error", 500)
			return
		}

		ts, err := template.ParseFiles([]string{"./templates/base.html", "./templates/home.html"}...)
		if err != nil {
			log.Println(err.Error())
			http.Error(w, "Internal Server Error", 500)
			return
		}

		err = ts.Execute(w, response{rgn, itunes.Regions, podcasts})
		if err != nil {
			log.Println(err.Error())
			http.Error(w, "Internal Server Error", 500)
			return
		}
	}
}

func (a *app) handleRegion() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			log.Println(err.Error())
			http.Error(w, "Internal Server Error", 500)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:  "region",
			Value: r.Form.Get("region"),
		})

		http.Redirect(w, r, "/", 301)
	}
}

func (a *app) handleSearch() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rgn := region(r)
		if err := r.ParseForm(); err != nil {
			log.Println(err.Error())
			http.Error(w, "Internal Server Error", 500)
			return
		}

		podcasts, err := a.str.Search(rgn, r.Form.Get("query"))
		if err != nil {
			log.Println(err.Error())
			http.Error(w, "Internal Server Error", 500)
			return
		}

		ts, err := template.ParseFiles([]string{"./templates/base.html", "./templates/results.html"}...)
		if err != nil {
			log.Println(err.Error())
			http.Error(w, "Internal Server Error", 500)
			return
		}

		err = ts.Execute(w, response{rgn, itunes.Regions, podcasts})
		if err != nil {
			log.Println(err.Error())
			http.Error(w, "Internal Server Error", 500)
			return
		}
	}
}

func (a *app) handlePodcast() http.HandlerFunc {
	type podcastAndReviews struct {
		Podcast *itunes.PodcastDetail
		Reviews []*itunes.Review
	}

	return func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]
		rgn := region(r)
		podcast, err := a.str.Lookup(id)
		if err != nil {
			log.Println(err.Error())
			http.Error(w, "Internal Server Error", 500)
			return
		}

		reviews, err := a.str.Reviews(id, rgn)
		if err != nil {
			log.Println(err.Error())
		}

		ts, err := template.ParseFiles([]string{"./templates/base.html", "./templates/podcast.html"}...)
		if err != nil {
			log.Println(err.Error())
			http.Error(w, "Internal Server Error", 500)
			return
		}

		err = ts.Execute(w, response{rgn, itunes.Regions, podcastAndReviews{podcast, reviews}})
		if err != nil {
			log.Println(err.Error())
			http.Error(w, "Internal Server Error", 500)
			return
		}
	}
}

func region(r *http.Request) string {
	cookie, err := r.Cookie("region")
	if err != nil {
		return ""
	}

	return cookie.Value
}

type response struct {
	Region  string
	Regions []itunes.Region
	Data    interface{}
}
