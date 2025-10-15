package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/joeblew999/infra/pkg/datastarui/sampleapp/pages"
)

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Handle("/css/*", http.StripPrefix("/css/", http.FileServer(http.Dir("static/css"))))
	r.Handle("/tests/*", http.StripPrefix("/tests/", http.FileServer(http.Dir("tests"))))

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	})

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		if err := pages.Home().Render(r.Context(), w); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	r.Get("/dashboard", func(w http.ResponseWriter, r *http.Request) {
		if err := pages.Dashboard().Render(r.Context(), w); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	log.Println("Sample app listening on http://localhost:4242")
	log.Fatal(http.ListenAndServe(":4242", r))
}
