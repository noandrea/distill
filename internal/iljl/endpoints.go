package iljl

import (
	"fmt"
	"net/http"
	"time"

	"gitlab.com/lowgroundandbigshoes/iljl/internal"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/render"
)

// RegisterEndpoints register application endpoints
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
	router.Use(middleware.Timeout(60 * time.Second))

	// redirect root to the configured url
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, internal.Config.Server.RootRedirect, 302)
	})
	// shortener redirect
	router.Get("/{ID}", func(w http.ResponseWriter, r *http.Request) {
		shortID := chi.URLParam(r, "ID")
		longURL, exists := fmt.Sprint(internal.Config.ShortID.Domain, "/", shortID), false
		// if the id does not existsk, send 404
		if !exists {
			http.Error(w, "URL not found", 404)
			return
		}
		// send redirect
		http.Redirect(w, r, longURL, 302)
	})
	// handle api requests
	router.Route("/api", func(r chi.Router) {

		cors := cors.New(cors.Options{
			AllowedOrigins:   []string{"*"},
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
			AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
			AllowCredentials: true,
			MaxAge:           300, // Maximum value not ignored by any of major browsers
		})
		// register cors and apiContext middleware
		r.Use(cors.Handler, apiContext)
		// handle global statistics
		r.Get("/stats", func(w http.ResponseWriter, r *http.Request) {
			render.JSON(w, r, "ok")
		})
		// handle url setup
		r.Post("/short", func(w http.ResponseWriter, r *http.Request) {
			render.JSON(w, r, "ok")
		})

		// delete an id
		r.Delete("/short/{ID}", func(w http.ResponseWriter, r *http.Request) {
			shortID := chi.URLParam(r, "ID")
			render.JSON(w, r, shortID)
		})
	})

	return router
}

// apiContext verify the api key header
func apiContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("X-API-KEY")
		if apiKey != internal.Config.Server.APIKey {
			http.Error(w, http.StatusText(403), 403)
			return
		}
		next.ServeHTTP(w, r)
	})
}
