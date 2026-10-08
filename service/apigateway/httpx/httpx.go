// Package httpx provides small helpers shared by the chi-based handlers:
// JSON response writing, request body binding, client IP extraction, and
// request-scoped values that were previously stored on echo.Context.
package httpx

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	sharederrors "github.com/MamangRust/microservice-ecommerce-shared/errors"
)

// ErrorResponse is the JSON shape for API error responses. It replaces the
// former sharederrors.ErrorResponse that was removed during the echo cleanup.
type ErrorResponse struct {
	Status      string                       `json:"status"`
	Message     string                       `json:"message"`
	Type        sharederrors.ErrorType       `json:"type,omitempty"`
	Code        int                          `json:"code"`
	TraceID     string                       `json:"trace_id,omitempty"`
	Retryable   bool                         `json:"retryable,omitempty"`
	Validations []sharederrors.ValidationError `json:"validations,omitempty"`
}


// JSON writes v as a JSON-encoded response with the given status code.
func JSON(w http.ResponseWriter, code int, v interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	return json.NewEncoder(w).Encode(v)
}

// Bind decodes the JSON request body into v.
func Bind(r *http.Request, v interface{}) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(v)
}

// RealIP resolves the client IP, honoring the same proxy headers as the
// previous echo implementation.
func RealIP(r *http.Request) string {
	if fw := r.Header.Get("X-Forwarded-For"); fw != "" {
		if i := strings.IndexByte(fw, ','); i >= 0 {
			fw = fw[:i]
		}
		fw = strings.TrimSpace(fw)
		if fw != "" {
			return fw
		}
	}

	if ip := strings.TrimSpace(r.Header.Get("X-Real-Ip")); ip != "" {
		return ip
	}

	if ip, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return ip
	}
	return r.RemoteAddr
}

// SetValue returns a copy of ctx carrying a request-scoped value. Keys are
// plain strings so handlers can keep addressing them by the same names that
// were used on echo.Context ("userId", "user_id", "role_names", ...).
func SetValue(ctx context.Context, key string, value interface{}) context.Context {
	return context.WithValue(ctx, key, value)
}

// Get reads a request-scoped value previously stored with SetValue.
func Get(r *http.Request, key string) interface{} {
	return r.Context().Value(key)
}

// Handler adapts an error-returning handler to http.HandlerFunc. It reproduces
// the former RegisterErrorHandler semantics: shared AppError values map to
// their own status code and JSON shape (trace_id from X-Request-ID), while any
// other error falls back to echo's default {"message": ...} format. Routes
// wrapped by the ApiHandler handle their own errors before returning.
func Handler(h func(http.ResponseWriter, *http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := h(w, r); err != nil {
			HandleHTTPError(w, r, err)
		}
	}
}

// HandleHTTPError writes err as a JSON error response following the former
// RegisterErrorHandler mapping.
func HandleHTTPError(w http.ResponseWriter, r *http.Request, err error) {
	var apiErr *sharederrors.AppError
	if errors.As(err, &apiErr) {
		_ = JSON(w, apiErr.Code, ErrorResponse{
			Status:      "error",
			Message:     apiErr.Message,
			Type:        apiErr.Type,
			Code:        apiErr.Code,
			TraceID:     r.Header.Get("X-Request-ID"),
			Retryable:   apiErr.Retryable,
			Validations: apiErr.Validations,
		})
		return
	}

	var httpErr *HTTPError
	if errors.As(err, &httpErr) {
		_ = JSON(w, httpErr.Code, map[string]string{"message": httpErr.Message})
		return
	}

	_ = JSON(w, http.StatusInternalServerError, map[string]string{"message": "Internal Server Error"})
}

// HTTPError is an error carrying a status code and message, written as JSON by
// WriteHTTPError. It replaces echo.NewHTTPError in middleware rejections.
type HTTPError struct {
	Code    int
	Message string
}

func (e *HTTPError) Error() string { return e.Message }

// NewHTTPError builds an *HTTPError with the given status code and message.
func NewHTTPError(code int, message string) error {
	return &HTTPError{Code: code, Message: message}
}

// WriteHTTPError writes err as a JSON error response. Middleware rejections
// previously flowed through echo's default handler, which writes the compact
// {"message": ...} shape; non-*HTTPError errors fall back to a generic 500.
func WriteHTTPError(w http.ResponseWriter, err error) {
	var httpErr *HTTPError
	if errors.As(err, &httpErr) {
		_ = JSON(w, httpErr.Code, map[string]string{"message": httpErr.Message})
		return
	}

	_ = JSON(w, http.StatusInternalServerError, map[string]string{"message": "Internal Server Error"})
}

// BindForm binds multipart or urlencoded form values into v using the `form`
// struct tags, mirroring echo's default binder for multipart requests.
func BindForm(r *http.Request, v interface{}) error {
	if err := r.ParseMultipartForm(32 << 20); err != nil && err != http.ErrNotMultipart {
		return err
	}

	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.Elem().Kind() != reflect.Struct {
		return errors.New("BindForm target must be a pointer to struct")
	}
	structVal := rv.Elem()
	structType := structVal.Type()

	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		if !field.IsExported() {
			continue
		}
		key := field.Tag.Get("form")
		if key == "" || key == "-" {
			continue
		}
		formVal := r.FormValue(key)
		if formVal == "" {
			continue
		}

		target := structVal.Field(i)
		if target.Kind() == reflect.Ptr {
			if target.IsNil() {
				target.Set(reflect.New(target.Type().Elem()))
			}
			target = target.Elem()
		}

		switch target.Kind() {
		case reflect.String:
			target.SetString(formVal)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if n, err := strconv.ParseInt(formVal, 10, 64); err == nil {
				target.SetInt(n)
			}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			if n, err := strconv.ParseUint(formVal, 10, 64); err == nil {
				target.SetUint(n)
			}
		case reflect.Float32, reflect.Float64:
			if fl, err := strconv.ParseFloat(formVal, 64); err == nil {
				target.SetFloat(fl)
			}
		case reflect.Bool:
			if b, err := strconv.ParseBool(formVal); err == nil {
				target.SetBool(b)
			}
		}
	}
	return nil
}
