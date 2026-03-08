package httpx

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/vgartg/goauction/internal/metrics"
)

func AccessLog(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			start := time.Now()
			defer func() {
				route := chi.RouteContext(r.Context()).RoutePattern()
				if route == "" {
					route = "unknown"
				}
				dur := time.Since(start)
				status := ww.Status()
				if status == 0 {
					status = http.StatusOK
				}
				metrics.HTTPRequestDuration.
					WithLabelValues(route, r.Method, strconv.Itoa(status)).
					Observe(dur.Seconds())
				logger.LogAttrs(r.Context(), slog.LevelInfo, "http",
					slog.String("request_id", middleware.GetReqID(r.Context())),
					slog.String("method", r.Method),
					slog.String("route", route),
					slog.String("path", r.URL.Path),
					slog.Int("status", status),
					slog.Int("bytes", ww.BytesWritten()),
					slog.Duration("duration", dur),
					slog.String("remote", r.RemoteAddr),
				)
			}()
			next.ServeHTTP(ww, r)
		})
	}
}
