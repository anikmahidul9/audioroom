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
		// ==========================================
		// 1. PUBLIC ROUTES (Auth ছাড়া অ্যাক্সেস করা যাবে)
		// ==========================================
		r.Get("/health", app.healthCheckHandler)
		r.Post("/auth/register", app.registerHandler) // Auth রাউট
		r.Post("/auth/login", app.loginHandler)       // Auth রাউট

		// ==========================================
		// 2. PROTECTED ROUTES (Auth Required)
		// ==========================================
		r.Group(func(r chi.Router) {
			r.Use(app.authMiddleware) // <-- Auth মিডলওয়্যার এখানে যোগ হলো

			r.Get("/me", func(w http.ResponseWriter, r *http.Request) {
				userID := r.Context().Value("userID").(string)
				app.jsonResponse(w, http.StatusOK, map[string]string{"userID": userID})
			})

			r.Route("/events", func(r chi.Router) {
				r.Post("/", app.createEventHandler)
				r.Get("/", app.feedEventHandler)

				r.Route("/{id}", func(r chi.Router) {
					r.Get("/", app.getEventHandler)
					r.Post("/start", app.startEventHandler)
					r.Post("/end", app.endEventHandler)

					r.Post("/join", app.joinRoomHandler)
					r.Post("/leave", app.leaveEventHandler)
					r.Get("/members", app.listMembersHandler)
					r.Patch("/members/{userID}/role", app.updateMemberRoleHandler)

					r.Post("/mute-all", app.muteAllSpeakersHandler)
					r.Post("/unmute-all", app.unmuteAllSpeakersHandler)
					r.Post("/members/{userID}/mute", app.forceMuteHandler)
					r.Post("/members/{userID}/kick", app.kickParticipantHandler)
				})
			})
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
