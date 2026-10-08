package middlewares

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
)

// TraceMiddleware extracts trace context from incoming HTTP headers and starts
// a server span for each request, enabling end-to-end distributed tracing from
// the API gateway down to the gRPC dependency services.
func TraceMiddleware(serviceName string) func(http.Handler) http.Handler {
	tracer := otel.Tracer(serviceName)
	propagator := otel.GetTextMapPropagator()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			req := r

			ctx := propagator.Extract(req.Context(), propagation.HeaderCarrier(req.Header))

			route := req.URL.Path
			if rctx := chi.RouteContext(ctx); rctx != nil {
				if pattern := rctx.RoutePattern(); pattern != "" {
					route = pattern
				}
			}
			spanName := req.Method + " " + route

			ctx, span := tracer.Start(ctx, spanName)
			defer span.End()

			span.SetAttributes(
				attribute.String("http.method", req.Method),
				attribute.String("http.route", route),
				attribute.String("http.target", req.URL.Path),
				attribute.String("http.host", req.Host),
			)

			// Propagate the traced context into the request so downstream
			// gRPC client interceptors can continue the same trace.
			req = req.WithContext(ctx)

			ww := newWrapResponseWriter(w)
			next.ServeHTTP(ww, req)

			span.SetAttributes(attribute.Int("http.status_code", ww.Status()))
			if ww.Status() >= 500 {
				span.SetStatus(codes.Error, "server error")
			}
		})
	}
}
