package gotelem

// this file defines the HTTP handlers and routes.

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/kschamplin/gotelem/skylab"
	"golang.org/x/exp/slog"
	"nhooyr.io/websocket"
)


type slogHttpLogger struct {
	slog.Logger
}

func TelemRouter(log *slog.Logger, broker *JBroker, db *TelemDb) http.Handler {
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

	r.Mount("/api/v1", apiV1(broker, db))

	// To future residents - you can add new API calls/systems in /api/v2
	// Don't break anything in api v1! keep legacy code working!

	// serve up a local status page.

	return r
}

// define API version 1 routes.
func apiV1(broker *JBroker, db *TelemDb) chi.Router {
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

	r.Route("/packets", func(r chi.Router) {
		r.Get("/subscribe", apiV1PacketSubscribe(broker, db))
		r.Post("/", func(w http.ResponseWriter, r *http.Request) {
			var pkgs []skylab.BusEvent
			decoder := json.NewDecoder(r.Body)
			if err := decoder.Decode(&pkgs); err != nil{
				w.WriteHeader(http.StatusTeapot)
				return
			}
			// we have a list of packets now. let's commit them.
			db.AddEvents(pkgs...)
			return 
		})
	})



	return r
}


// apiV1Subscriber is a websocket session for the v1 api.
type apiV1Subscriber struct {
	idFilter []uint64 // list of Ids to subscribe to. If it's empty, subscribes to all.
}

func apiV1PacketSubscribe(broker *JBroker, db *TelemDb) http.HandlerFunc {
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
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "error ws handshake: %s", err)
			return
		}

		sess := &apiV1Subscriber{}

		for {
			select {
			case <- r.Context().Done():
				return
			case msgIn := <-sub:
				if len(sess.idFilter) == 0 {
					// send it.
					goto escapeFilter
				}
				for _, id := range sess.idFilter {
					if id == msgIn.Id {
						// send it
					}
				}
				escapeFilter:
				return

			}


		}
		


	}
}

