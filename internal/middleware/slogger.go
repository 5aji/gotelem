package middleware

import (
	"context"
	"net/http"
	"time"

	"log/slog"

	chi_middleware "github.com/go-chi/chi/v5/middleware"
)

// Slogger is a slog-enabled logging middleware.
// It logs the start and end of the request, and logs info
// about the request itself, response status, and response time.

// Slogger returns a log handler that uses the given slog logger as the base.
func Slogger(sl *slog.Logger) func(next http.Handler) http.Handler {

	logger := sl.WithGroup("http")
	return func(next http.Handler) http.Handler {

		// this triple-nested function is strange, but basically the Slogger() call makes a new middleware function (above)
		// the middleware function returns a handler that calls the next handler in the chain(wrapping it)

		fn := func(w http.ResponseWriter, r *http.Request) {
			// wrap writer allows us to get info on the response from further handlers.
			ww := chi_middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			t1 := time.Now()
			// attrs is stored to allow for the helpers to add additional elements to the main record.
			attrs := make([]slog.Attr, 0)

			// This function runs at the end and adds all the response details to the attrs before logging them.
			defer func() {
				attrs = append(attrs, slog.Int("status_code", ww.Status()))
				attrs = append(attrs, slog.Int("resp_size", ww.BytesWritten()))
				attrs = append(attrs, slog.Duration("duration", time.Since(t1)))
				attrs = append(attrs, slog.String("method", r.Method))
				logger.LogAttrs(r.Context(), slog.LevelInfo, r.RequestURI, attrs...)

			}()

			// embed the logger and the attrs for later items in the chain.
			ctx := context.WithValue(r.Context(), SloggerAttrsKey, attrs)
			ctx = context.WithValue(ctx, SloggerLogKey, logger)
			// push it to the request and serve the next handler
			r = r.WithContext(ctx)
			next.ServeHTTP(ww, r)
		}

		return http.HandlerFunc(fn)
	}
}

type slogKeyType int

const (
	SloggerLogKey slogKeyType = iota
	SloggerAttrsKey
)

func AddSlogAttr(r *http.Request, attr slog.Attr) {
	ctx := r.Context()
	attrs, ok := ctx.Value(SloggerAttrsKey).([]slog.Attr)
	if !ok {
		return
	}
	attrs = append(attrs, attr)

}

// TODO: write rest of functions
