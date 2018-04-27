package iljl

import (
	"net/http"
	"time"

	"gitlab.com/lowgroundandbigshoes/iljl/internal"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/unrolled/render"
)

func RegisterEndpoints() (router *chi.Mux) {
	router = chi.NewRouter()

	// A good base middleware stack
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	// Set a timeout value on the request context (ctx), that will signal
	// through ctx.Done() that the request has timed out and further
	// processing should be stopped.
	r.Use(middleware.Timeout(60 * time.Second))

	// redirect the root to the configured url
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, internal.Config.Server.RootRedirect, 302)
	})
	// shortener redirect
	r.Get("/{ID}", func(w http.ResponseWriter, r *http.Request) {
		shortID := chi.URLParam(r, "ID")
		longURL, exists := "string", false
		if !exists {
			http.Error(w, "URL not found", 404)
			return
		}
		http.Redirect(w, r, longURL, 302)
	})
	// handle api requests
	router.Route("/api", func(r chi.Router) {
		r.Get("/stats", getGlobalStatistics) // GET /articles
		r.Post("/short", addURL)             // GET /articles/01-16-2017
		r.Delete("/short/{ID}", func(w http.ResponseWriter, r *http.Request) {
			shortID := chi.URLParam(r, "ID")
			render.JSON(w, 200, "ok")
		})
	})

	router

	return router
}
