// Package web provides the api endpoints for the urstore
package web

import (
	"crypto/subtle"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/noandrea/distill/config"
	"github.com/noandrea/distill/pkg/distill"
	"github.com/noandrea/distill/pkg/model"
	log "github.com/sirupsen/logrus"
)

var (
	settings config.Schema
)

// RegisterEndpoints register application endpoints
func RegisterEndpoints(cfg config.Schema) (router *chi.Mux) {
	router = chi.NewRouter()
	settings = cfg

	// A good base middleware stack
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	// Set a timeout value on the request context (ctx), that will signal
	// through ctx.Done() that the request has timed out and further
	// processing should be stopped.
	router.Use(middleware.Timeout(60 * time.Second))
	// register cors and apiContext middleware
	router.Use(cors)
	// profiler route
	router.Mount("/profile", middleware.Profiler())

	// health check route
	router.Get("/health-check", healthCheckHandler)
	// redirect root to the configured url
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, settings.ShortID.RootRedirectURL, 302)
	})
	// shortener redirect
	router.Get("/{ID}", handleGetURL)
	// handle api requests
	router.Route("/api", func(r chi.Router) {
		r.Use(apiContext)
		// handle global statistics
		r.Get("/stats", handleGetStats)
		r.Delete("/stats", handleResetStats)
		// handle url statistics
		r.Get("/stats/{ID}", handleStatsURL)
		// handle url setup
		r.Post("/short", handleShort)
		// implement kutt.it endpoint
		r.Post("/url/submit", handleShort)
		// delete an id
		r.Delete("/short/{ID}", handleDeleteURL)
		// backup
	})
	return router
}

//   ____  ____       _       ____  _____  ______   _____     ________  _______     ______
//  |_   ||   _|     / \     |_   \|_   _||_   _ `.|_   _|   |_   __  ||_   __ \  .' ____ \
//    | |__| |      / _ \      |   \ | |    | | `. \ | |       | |_ \_|  | |__) | | (___ \_|
//    |  __  |     / ___ \     | |\ \| |    | |  | | | |   _   |  _| _   |  __ /   _.____`.
//   _| |  | |_  _/ /   \ \_  _| |_\   |_  _| |_.' /_| |__/ | _| |__/ | _| |  \ \_| \____) |
//  |____||____||____| |____||_____|\____||______.'|________||________||____| |___|\______.'
//

func handleShort(w http.ResponseWriter, r *http.Request) {
	urlReq := &model.URLReq{}
	if err := render.Bind(r, urlReq); err != nil {
		render.Render(w, r, ErrInvalidRequest(err, err.Error()))
		return
	}
	// retrieve the forceAlphabet and forceLength
	forceAlphabet, forceLength := false, false
	fA := chi.URLParam(r, "forceAlphabet")
	fL := chi.URLParam(r, "forceLength")
	if fA == "1" {
		forceAlphabet = true
	}
	if fL == "1" {
		forceLength = true
	}
	// upsert the data
	id, err := distill.UpsertURL(urlReq, forceAlphabet, forceLength)
	log.Trace("created %v", id)
	// TODO: check the actual error
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err, err.Error()))
		return
	}
	render.JSON(w, r, map[string]string{"ID": id})
}

func handleGetURL(w http.ResponseWriter, r *http.Request) {
	shortID := chi.URLParam(r, "ID")
	targetURL, err := distill.GetURLRedirect(shortID)
	if err != nil && len(targetURL) == 0 {
		http.Error(w, "URL not found", 404)
	}
	// send redirect
	http.Redirect(w, r, targetURL, 302)
	return
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("good"))
}

func handleStatsURL(w http.ResponseWriter, r *http.Request) {
	shortID := chi.URLParam(r, "ID")
	if urlInfo, err := distill.GetURLInfo(shortID); err == nil {
		// send redirect
		render.JSON(w, r, urlInfo)
		return
	}
	http.Error(w, "URL not found", 404)
}

func handleGetStats(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, map[string]string{})
}

func handleResetStats(w http.ResponseWriter, r *http.Request) {
	// err := urlstore.ResetStats()
	// if err != nil {
	// 	render.Render(w, r, ErrInternalError(err, err.Error()))
	// 	return
	// }
	render.JSON(w, r, map[string]string{})
}

func handleDeleteURL(w http.ResponseWriter, r *http.Request) {
	shortID := chi.URLParam(r, "ID")
	err := distill.DeleteURL(shortID)
	if err != nil {
		render.Render(w, r, ErrNotFound(err, "URL id not found"))
		return
	}
	render.JSON(w, r, map[string]string{"ID": shortID})
}

//   ____    ____   ______     ______    ______
//  |_   \  /   _|.' ____ \  .' ___  | .' ____ \
//    |   \/   |  | (___ \_|/ .'   \_| | (___ \_|
//    | |\  /| |   _.____`. | |   ____  _.____`.
//   _| |_\/_| |_ | \____) |\ `.___]  || \____) |
//  |_____||_____| \______.' `._____.'  \______.'
//

// ErrInvalidRequest render an invalid request
func ErrInvalidRequest(err error, message string) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: http.StatusBadRequest,
		AppCode:        http.StatusBadRequest,
		ErrorText:      message,
	}
}

// ErrInternalError render an invalid request
func ErrInternalError(err error, message string) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: http.StatusInternalServerError,
		AppCode:        http.StatusInternalServerError,
		ErrorText:      message,
	}
}

// ErrNotFound render an invalid request
func ErrNotFound(err error, message string) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: http.StatusNotFound,
		AppCode:        http.StatusNotFound,
		ErrorText:      message,
	}
}

// Render an ErrResponse
func (e *ErrResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPStatusCode)
	return nil
}

// ErrResponse renderer type for handling all sorts of errors.
//
// In the best case scenario, the excellent github.com/pkg/errors package
// helps reveal information on the error, setting it on Err, and in the Render()
// method, using it to set the application-specific error code in AppCode.
type ErrResponse struct {
	Err            error `json:"-"` // low-level runtime error
	HTTPStatusCode int   `json:"-"` // http response status code

	AppCode   int    `json:"code,omitempty"`    // application-specific error code
	ErrorText string `json:"message,omitempty"` // application-level error message, for debugging
}

//   ____    ____  _____  ______   ____      ____  _       _______     ________
//  |_   \  /   _||_   _||_   _ `.|_  _|    |_  _|/ \     |_   __ \   |_   __  |
//    |   \/   |    | |    | | `. \ \ \  /\  / / / _ \      | |__) |    | |_ \_|
//    | |\  /| |    | |    | |  | |  \ \/  \/ / / ___ \     |  __ /     |  _| _
//   _| |_\/_| |_  _| |_  _| |_.' /_  \  /\  /_/ /   \ \_  _| |  \ \_  _| |__/ |
//  |_____||_____||_____||______.'(_)  \/  \/|____| |____||____| |___||________|
//

// apiContext verify the api key header
func apiContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get(settings.Tuning.APIKeyHeaderName)
		keyMatch := subtle.ConstantTimeCompare([]byte(apiKey), []byte(settings.Server.APIKey))
		if keyMatch == 0 {
			http.Error(w, http.StatusText(403), 403)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// cors handler for cors headers
func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "*")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == "OPTIONS" {
			log.Printf("Should return for OPTIONS")
			return
		}
		next.ServeHTTP(w, r)
	})
}
