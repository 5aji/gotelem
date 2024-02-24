package gotelem

// this file defines the HTTP handlers and routes.

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/kschamplin/gotelem/internal/db"
	"github.com/kschamplin/gotelem/skylab"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

type slogHttpLogger struct {
	slog.Logger
}

func TelemRouter(log *slog.Logger, broker *Broker, db *db.TelemDb) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger) // TODO: integrate with slog instead of go default logger.
	r.Use(middleware.Recoverer)

	// heartbeat request.
	r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	})

	r.Mount("/api/v1", apiV1(broker, db))

	// To future residents - you can add new API calls/systems in /api/v2
	// Don't break anything in api v1! keep legacy code working!

	return r
}

// define API version 1 routes.
func apiV1(broker *Broker, db *db.TelemDb) chi.Router {
	r := chi.NewRouter()
	// this API only accepts JSON.
	r.Use(middleware.AllowContentType("application/json"))
	// no caching - always get the latest data.
	// TODO: add a smart short expiry cache for queries that take a while.
	r.Use(middleware.NoCache)

	r.Get("/schema", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// return the Skylab JSON definitions
		w.Write([]byte(skylab.SkylabDefinitions))
	})

	r.Route("/packets", func(r chi.Router) {
		r.Get("/subscribe", apiV1PacketSubscribe(broker, db))
		r.Post("/", func(w http.ResponseWriter, r *http.Request) {
			var pkgs []skylab.BusEvent
			decoder := json.NewDecoder(r.Body)
			if err := decoder.Decode(&pkgs); err != nil {
				w.WriteHeader(http.StatusTeapot)
				return
			}
			// we have a list of packets now. let's commit them.
			db.AddEvents(pkgs...)
		})
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			// this should use http query params to return a list of packets.

		})

		// this is to get a single field
		r.Get("/{name:[a-z_]+}/{field:[a-z_]+}")

	})

	// records are driving segments/runs.
	r.Route("/records", func(r chi.Router) {
		r.Get("/", apiV1GetRecords(db))            // get all runs
		r.Get("/active", apiV1GetActiveRecord(db)) // get current run (no end time)
		r.Post("/", apiV1StartRecord(db))          // create a new run (with note). Ends active run if any, and creates new active run (no end time)
		r.Get("/{id}", apiV1GetRecord(db))         // get details on a specific run
		r.Put("/{id}", apiV1UpdateRecord(db))      // update a specific run. Can only be used to add notes/metadata, and not to change time/id.

	})

	r.Get("/stats", func(w http.ResponseWriter, r *http.Request) {

	}) // v1 api stats (calls, clients, xbee connected, meta health ok)

	return r
}

// apiV1Subscriber is a websocket session for the v1 api.
type apiV1Subscriber struct {
	nameFilter []string // names of packets we care about.
}

// this is a websocket stream.
func apiV1PacketSubscribe(broker *Broker, db *db.TelemDb) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn_id := r.RemoteAddr + uuid.New().String()
		sub, err := broker.Subscribe(conn_id)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "error subscribing: %s", err)
			return
		}
		defer broker.Unsubscribe(conn_id)
		// attempt to upgrade.
		c, err := websocket.Accept(w, r, nil)
		c.Ping(r.Context())
		if err != nil {
			// TODO: is this the correct option?
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "error ws handshake: %s", err)
			return
		}

		// TODO: use K/V with session token?
		sess := &apiV1Subscriber{}

		for {
			select {
			case <-r.Context().Done():
				return
			case msgIn := <-sub:
				if len(sess.nameFilter) == 0 {
					// send it.
					wsjson.Write(r.Context(), c, msgIn)
				}
				for _, name := range sess.nameFilter {
					if name == msgIn.Name {
						// send it
						wsjson.Write(r.Context(), c, msgIn)
						break
					}
				}

			}

		}

	}
}

// apiV1GetValues is a function that creates a handler for
// getting the specific value from a packet.
// this is useful for OpenMCT or other viewer APIs
func apiV1GetValues(db *db.TelemDb) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error

		// we need a start and end time. If none is provided,
		// we use unix epoch as start, and now + 1 day as end.
		start := time.Unix(0, 0)
		startString := r.URL.Query().Get("start")
		if startString != "" {
			start, err = time.Parse(time.RFC3339, startString)
			if err != nil {

			}
		}
		end := time.Now().Add(1 * time.Hour)
		endParam := r.URL.Query().Get("start")
		if endParam != "" {
			end, err = time.Parse(time.RFC3339, endParam)
			if err != nil {
			}
		}
		name := chi.URLParam(r, "name")
		field := chi.URLParam(r, "field")
		
		// TODO: add limit and pagination

		res, err := db.GetValues(r.Context(), name, field, start, end)
		if err != nil {
			// 500 server error:
			fmt.Print(err)
		}
		b, err := json.Marshal(res)
		w.Write(b)
	}

}

// TODO: rename. record is not a clear name. Runs? drives? segments?
func apiV1GetRecords(db *db.TelemDb) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}

func apiV1GetActiveRecord(db *db.TelemDb) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}

func apiV1StartRecord(db *db.TelemDb) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {}
}

func apiV1GetRecord(db *db.TelemDb) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {}
}

func apiV1UpdateRecord(db *db.TelemDb) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {}
}
