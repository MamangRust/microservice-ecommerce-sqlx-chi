package middlewares

import "net/http"

// wrapResponseWriter records the status code written by the handler so
// observability middlewares (metrics, tracing) can report it, mirroring the
// c.Response().Status access that echo provided.
type wrapResponseWriter struct {
	http.ResponseWriter
	status int
}

func newWrapResponseWriter(w http.ResponseWriter) *wrapResponseWriter {
	return &wrapResponseWriter{ResponseWriter: w, status: http.StatusOK}
}

func (ww *wrapResponseWriter) WriteHeader(code int) {
	ww.status = code
	ww.ResponseWriter.WriteHeader(code)
}

// Flush implements http.Flusher for streaming handlers.
func (ww *wrapResponseWriter) Flush() {
	if f, ok := ww.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func (ww *wrapResponseWriter) Status() int { return ww.status }
