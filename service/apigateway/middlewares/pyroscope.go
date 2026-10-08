package middlewares

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/grafana/pyroscope-go"
)

func PyroscopeMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			endpoint := r.URL.Path
			if rctx := chi.RouteContext(r.Context()); rctx != nil {
				if pattern := rctx.RoutePattern(); pattern != "" {
					endpoint = pattern
				}
			}
			pyroscope.TagWrapper(r.Context(), pyroscope.Labels(
				"path", endpoint,
				"method", r.Method,
			), func(ctx context.Context) {
				// TagWrapper replaces the context; the request must carry it so
				// downstream handlers stay inside the tagged profiling span.
				next.ServeHTTP(w, r.WithContext(ctx))
			})
		})
	}
}
