package gotelem

// this file defines the HTTP handlers and routes.

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/kschamplin/gotelem/skylab"
	"golang.org/x/exp/slog"
	"nhooyr.io/websocket"
)

type slogHttpLogger struct {
	slog.Logger
}

func TelemRouter(log *slog.Logger, broker *JBroker) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Logger) // TODO: integrate with slog
	r.Use(middleware.Recoverer)

	r.Get("/schema", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// return the spicy json response.
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(skylab.SkylabDefinitions))
	})

	// heartbeat request.
	r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	})

	r.Mount("/api/v1", apiV1(broker))

	// To future residents - you can add new API calls/systems in /api/v2
	// Don't break anything in api v1! keep legacy code working!

	// serve up a local status page.

	return r
}

// define API version 1 routes.
func apiV1(broker *JBroker) chi.Router {
	r := chi.NewRouter()
	r.Get("/schema", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// return the spicy json response.
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(skylab.SkylabDefinitions))
	})

	r.Get("/ws", func(w http.ResponseWriter, r *http.Request) {
		c, err := websocket.Accept(w, r, nil)
		if err != nil {
			return
		}
		
	})

	return r
}

