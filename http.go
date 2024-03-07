package gotelem

// this file defines the HTTP handlers and routes.

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/kschamplin/gotelem/skylab"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

func extractBusEventFilter(r *http.Request) (*BusEventFilter, error) {

	bef := &BusEventFilter{}

	v := r.URL.Query()
	bef.Names = v["name"] // put all the names in.
	if el := v.Get("start"); el != "" {
		// parse the start time query.
		t, err := time.Parse(time.RFC3339, el)
		if err != nil {
			return bef, err
		}
		bef.StartTime = t
	}
	if el := v.Get("end"); el != "" {
		// parse the start time query.
		t, err := time.Parse(time.RFC3339, el)
		if err != nil {
			return bef, err
		}
		bef.EndTime = t
	}
	bef.Indexes = make([]int, 0)
	for _, strIdx := range v["idx"] {
		idx, err := strconv.ParseInt(strIdx, 10, 32)
		if err != nil {
			return nil, err
		}
		bef.Indexes = append(bef.Indexes, int(idx))
	}
	return bef, nil
}

func extractLimitModifier(r *http.Request) (*LimitOffsetModifier, error) {
	lim := &LimitOffsetModifier{}
	v := r.URL.Query()
	if el := v.Get("limit"); el != "" {
		val, err := strconv.ParseInt(el, 10, 64)
		if err != nil {
			return nil, err
		}
		lim.Limit = int(val)
		// next, we check if we have an offset.
		// we only check offset if we also have a limit.
		// offset without limit isn't valid and is ignored.
		if el := v.Get("offset"); el != "" {
			val, err := strconv.ParseInt(el, 10, 64)
			if err != nil {
				return nil, err
			}
			lim.Offset = int(val)
		}
		return lim, nil
	}
	// we use the nil case to indicate that no limit was provided.
	return nil, nil
}

type RouterMod func(chi.Router)

var RouterMods = []RouterMod{}

func TelemRouter(log *slog.Logger, broker *Broker, db *TelemDb) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger) // TODO: integrate with slog instead of go default logger.
	r.Use(middleware.Recoverer)
	r.Use(middleware.SetHeader("Access-Control-Allow-Origin", "*"))

	// heartbeat request.
	r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	})

	r.Mount("/api/v1", apiV1(broker, db))

	for _, mod := range RouterMods {
		mod(r)
	}
	// To future residents - you can add new API calls/systems in /api/v2
	// Don't break anything in api v1! keep legacy code working!

	return r
}

// define API version 1 routes.
func apiV1(broker *Broker, tdb *TelemDb) chi.Router {
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

	// OpenMCT domain object storage. Basically an arbitrary JSON document store

	r.Route("/openmct", func(r chi.Router) {
		// key is a column on our json store, it's nested under identifier.key
		r.Get("/{key}", func(w http.ResponseWriter, r *http.Request) {})
		r.Put("/{key}", func(w http.ResponseWriter, r *http.Request) {})
		r.Delete("/{key}", func(w http.ResponseWriter, r *http.Request) {})
		// create a new object.
		r.Post("/", func(w http.ResponseWriter, r *http.Request) {})
		// subscribe to object updates.
		r.Get("/subscribe", func(w http.ResponseWriter, r *http.Request) {})
	})

	// records are driving segments/runs.

	r.Get("/stats", func(w http.ResponseWriter, r *http.Request) {

	}) // v1 api stats (calls, clients, xbee connected, meta health ok)

	return r
}

// this is a websocket stream.
func apiV1PacketSubscribe(broker *Broker, db *TelemDb) http.HandlerFunc {
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

func apiV1GetPackets(tdb *TelemDb) http.HandlerFunc {
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
			fmt.Print(lim)
			return
		}

		// TODO: is the following check needed?
		var res []skylab.BusEvent
		res, err = tdb.GetPackets(r.Context(), *bef, lim)
		if err != nil {
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

// apiV1GetValues is a function that creates a handler for
// getting the specific value from a packet.
// this is useful for OpenMCT or other viewer APIs
func apiV1GetValues(db *TelemDb) http.HandlerFunc {
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

		var res []Datum
		// make the call, skip the limit modifier if it's nil.
		res, err = db.GetValues(r.Context(), *bef, field, lim)
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
