package main

import (
	"log"
	"net/http"
	"time"

	"github.com/anikmahidul9/voiceroom/internal/store"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type application struct {
	config config
	store  store.Storage
}

type dbConfig struct {
	addr         string
	maxOpenConns int
	maxIdleConns int
	maxIdleTime  string
}

type config struct {
	addr string
	db   dbConfig
	env  string
}

func (app *application) mount() *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(middleware.Timeout(time.Second * 60))

	r.Route("/v1", func(r chi.Router) {
		r.Get("/health", app.healthCheckHandler)
		r.Route("/events", func(r chi.Router) {
			// Base path: /v1/events
			r.Post("/", app.createEventHandler)
			r.Get("/", app.feedEventHandler)

			// Path: /v1/events/{id}
			r.Get("/{id}", app.getEventHandler)

			// Path: /v1/events/{id}/start
			r.Post("/{id}/start", app.startEventHandler)

			// Path: /v1/events/{id}/end
			r.Post("/{id}/end", app.endEventHandler)
			r.Post("/join", app.joinEventHandler)
			r.Post("/leave", app.leaveEventHandler)
			r.Get("/members", app.listMembersHandler)
			r.Patch("/members/{userID}/role", app.updateMemberRoleHandler)
		})
	})

	return r

}
func (app *application) run(mux *chi.Mux) error {
	srv := &http.Server{
		Addr:         app.config.addr,
		Handler:      mux,
		WriteTimeout: time.Second * 30,
		ReadTimeout:  time.Second * 30,
		IdleTimeout:  time.Second * 60,
	}

	log.Printf("Server is starting at %s", app.config.addr)

	return srv.ListenAndServe()
}
