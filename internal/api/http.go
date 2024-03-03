package api

// this file defines the HTTP handlers and routes.

import (
	"encoding/json"
	"fmt"
	"net/http"

	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/kschamplin/gotelem"
	"github.com/kschamplin/gotelem/internal/db"
	"github.com/kschamplin/gotelem/skylab"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

func TelemRouter(log *slog.Logger, broker *gotelem.Broker, db *db.TelemDb) http.Handler {
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
func apiV1(broker *gotelem.Broker, tdb *db.TelemDb) chi.Router {
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
		r.Get("/subscribe", apiV1PacketSubscribe(broker, tdb))
		r.Post("/", func(w http.ResponseWriter, r *http.Request) {
			var pkgs []skylab.BusEvent
			decoder := json.NewDecoder(r.Body)
			if err := decoder.Decode(&pkgs); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			tdb.AddEvents(pkgs...)
		})
		// general packet history get.
		r.Get("/", apiV1GetPackets(tdb))

		// this is to get a single field from a packet.
		r.Get("/{name:[a-z_]+}/{field:[a-z_]+}", apiV1GetValues(tdb))

	})

	// records are driving segments/runs.
	r.Route("/records", func(r chi.Router) {
		r.Get("/", apiV1GetRecords(tdb))            // get all runs
		r.Get("/active", apiV1GetActiveRecord(tdb)) // get current run (no end time)
		r.Post("/", apiV1StartRecord(tdb))          // create a new run (with note). Ends active run if any, and creates new active run (no end time)
		r.Get("/{id}", apiV1GetRecord(tdb))         // get details on a specific run
		r.Put("/{id}", apiV1UpdateRecord(tdb))      // update a specific run. Can only be used to add notes/metadata, and not to change time/id.

	})

	r.Get("/stats", func(w http.ResponseWriter, r *http.Request) {

	}) // v1 api stats (calls, clients, xbee connected, meta health ok)

	return r
}


// this is a websocket stream.
func apiV1PacketSubscribe(broker *gotelem.Broker, db *db.TelemDb) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// pull filter from url query params.
		bef, err := extractBusEventFilter(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		// setup connection
		conn_id := r.RemoteAddr + uuid.New().String()
		sub, err := broker.Subscribe(conn_id)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "error subscribing: %s", err)
			return
		}
		defer broker.Unsubscribe(conn_id)

		// setup websocket
		c, err := websocket.Accept(w, r, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		c.Ping(r.Context())
		// closeread handles protocol/status messages,
		// also handles clients closing the connection.
		// we get a context to use from it.
		ctx := c.CloseRead(r.Context())

		

		for {
			select {
			case <-ctx.Done():
				return
			case msgIn := <-sub:
				// short circuit if there's no names - send everything
				if len(bef.Names) == 0 {
					wsjson.Write(r.Context(), c, msgIn)
				}
				// otherwise, send it if it matches one of our names.
				for _, name := range bef.Names {
					if name == msgIn.Name {
						// send it
						wsjson.Write(ctx, c, msgIn)
						break
					}
				}

			}
		}

	}
}

func apiV1GetPackets(tdb *db.TelemDb) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// this should use http query params to return a list of packets.
		bef, err := extractBusEventFilter(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		lim, err := extractLimitModifier(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// TODO: is the following check needed?
		var res []skylab.BusEvent
		if lim != nil {
			res, err = tdb.GetPackets(r.Context(), *bef, lim)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

		} else {
			res, err = tdb.GetPackets(r.Context(), *bef)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		b, err := json.Marshal(res)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(b)

	}
}

// apiV1GetValues is a function that creates a handler for
// getting the specific value from a packet.
// this is useful for OpenMCT or other viewer APIs
func apiV1GetValues(db *db.TelemDb) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error

		bef, err := extractBusEventFilter(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		lim, err := extractLimitModifier(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// get the URL parameters, these are guaranteed to exist.
		name := chi.URLParam(r, "name")
		field := chi.URLParam(r, "field")

		// override the bus event filter name option
		bef.Names = []string{name}

		res, err := db.GetValues(r.Context(), *bef, field, lim)
		if err != nil {
			// 500 server error:
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		b, err := json.Marshal(res)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
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
