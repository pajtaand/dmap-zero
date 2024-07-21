package middleware

import (
	"net/http"
	"time"

	"github.com/andreepyro/dmap-zero/internal/common/constants"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog/log"
)

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{w, http.StatusNotImplemented}
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		ctx := r.Context()
		requestID := middleware.GetReqID(ctx)
		l := log.Logger.With().
			Str(constants.LoggerKeyRequestID, requestID).Logger()
		ctx = l.WithContext(ctx)
		r = r.WithContext(ctx)

		lrw := newLoggingResponseWriter(w)

		l.Info().
			Str(constants.LoggerKeyUrl, r.URL.RequestURI()).
			Str(constants.LoggerKeyMethod, r.Method).
			Str(constants.LoggerKeyUserAgent, r.UserAgent()).
			Msg("Incoming request")

		defer func() {
			l.Info().
				Dur(constants.LoggerKeyElapsedTime, time.Since(start)).
				Int(constants.LoggerKeyStatusCode, lrw.statusCode).
				Msg("Request response")

		}()

		next.ServeHTTP(lrw, r)
	})
}
