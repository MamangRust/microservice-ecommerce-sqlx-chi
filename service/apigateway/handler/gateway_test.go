package handler_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/MamangRust/microservice-ecommerce-grpc-apigateway/handler"
	"github.com/MamangRust/microservice-ecommerce-grpc-apigateway/httpx"
	"github.com/MamangRust/microservice-ecommerce-grpc-apigateway/middlewares"
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	httpSwagger "github.com/swaggo/http-swagger"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func testLogger(t *testing.T) logger.LoggerInterface {
	logger.ResetInstance()
	l, err := logger.NewLogger("apigateway-test", sdklog.NewLoggerProvider())
	require.NoError(t, err)
	require.NotNil(t, l)
	return l
}

// bootGateway mirrors the production wiring (apps/client.go createServer +
// handler.NewHandler): global JWT middleware with the real whitelist, swagger,
// health, and every registered domain handler. gRPC connections may be nil
// because handlers only use them at request time.
func bootGateway(t *testing.T, conns *handler.ServiceConnections) chi.Router {
	viper.Set("SECRET_KEY", "test-secret")
	r := chi.NewRouter()
	r.Use(middlewares.JWTAuth())
	r.Get("/swagger/*", httpSwagger.WrapHandler)
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		_ = httpx.JSON(w, http.StatusOK, map[string]string{"status": "healthy"})
	})
	handler.NewHandler(&handler.Deps{
		Router:             r,
		Logger:             testLogger(t),
		ServiceConnections: conns,
		Cache:              nil,
		Image:              nil,
		Kafka:              nil,
	})
	return r
}

type routeSet map[string]bool

func collectRoutes(r chi.Router) routeSet {
	set := routeSet{}
	chi.Walk(r, func(method, route string, _ http.Handler, _ ...func(http.Handler) http.Handler) error {
		// chi reports subrouter-root GETs with a trailing slash while the
		// swagger annotations (and the former echo route table) use none.
		set[method+" "+strings.TrimSuffix(route, "/")] = true
		return nil
	})
	return set
}

func hasRoute(routes routeSet, method, path string) bool {
	return routes[method+" "+path]
}

func splitRouteKey(key string) (string, string) {
	parts := strings.SplitN(key, " ", 2)
	if len(parts) != 2 {
		return "", key
	}
	return parts[0], parts[1]
}

func signToken(t *testing.T, sub string) string {
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": sub})
	signed, err := tok.SignedString([]byte("test-secret"))
	require.NoError(t, err)
	return signed
}

func perform(e chi.Router, method, path string, body io.Reader, token string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, body)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

// TestGatewayRouteInventory validates that the live route table, built from the
// actual handler registration, follows the -command/-query contract and has no
// swapped groups (regression for the cart route swap).
func TestGatewayRouteInventory(t *testing.T) {
	e := bootGateway(t, &handler.ServiceConnections{})

	routes := collectRoutes(e)
	require.NotEmpty(t, routes, "gateway must register routes")

	// Cart command/query separation (regression for the swapped groups).
	require.True(t, hasRoute(routes, http.MethodPost, "/api/cart-command/create"))
	require.True(t, hasRoute(routes, http.MethodDelete, "/api/cart-command/delete"))
	require.True(t, hasRoute(routes, http.MethodPost, "/api/cart-command/delete-all"))
	require.True(t, hasRoute(routes, http.MethodGet, "/api/cart-query"))
	require.False(t, hasRoute(routes, http.MethodPost, "/api/cart-query/create"),
		"command must not be registered under the cart-query group")

	// Convention: groups ending in -command only carry mutations; -query only GETs.
	for key := range routes {
		method, path := splitRouteKey(key)
		api := strings.TrimPrefix(path, "/api/")
		group := strings.SplitN(api, "/", 2)[0]
		switch {
		case strings.HasSuffix(group, "-command"):
			require.NotEqual(t, http.MethodGet, method, "GET must not be a command route: %s", path)
		case strings.HasSuffix(group, "-query"):
			require.Equal(t, http.MethodGet, method, "only GET allowed on query group: %s", path)
		}
	}

	require.True(t, hasRoute(routes, http.MethodGet, "/health"))
	require.True(t, hasRoute(routes, http.MethodGet, "/swagger/*"))
}

// TestSwaggerMatchesRegisteredRoutes enforces that the generated Swagger spec
// documents every route actually registered by the gateway (and nothing else of
// significance). Chi :params are normalized to Swagger {params} before the set
// comparison; regenerate with `just generate-swagger` after route changes.
func TestSwaggerMatchesRegisteredRoutes(t *testing.T) {
	e := bootGateway(t, &handler.ServiceConnections{})

	raw, err := os.ReadFile("../docs/swagger.json")
	require.NoError(t, err, "run `just generate-swagger` to produce docs/swagger.json")

	var spec struct {
		Paths map[string]json.RawMessage `json:"paths"`
	}
	require.NoError(t, json.Unmarshal(raw, &spec))

	paramRegex := regexp.MustCompile(`:([a-zA-Z_]+)`)
	var missing []string
	for key := range collectRoutes(e) {
		method, path := splitRouteKey(key)
		if !strings.HasPrefix(path, "/api/") {
			continue
		}
		swaggerPath := paramRegex.ReplaceAllString(path, "{$1}")
		if _, ok := spec.Paths[swaggerPath]; !ok {
			missing = append(missing, method+" "+path)
		}
	}
	require.Empty(t, missing,
		"registered routes missing from swagger.json (regenerate docs and fix annotations):\n%s",
		strings.Join(missing, "\n"))
}

// TestGatewayMiddlewareSmoke covers the exit-criteria status matrix on the real
// middleware chain: health 200 without token, 401, 404, 400 (validation), and
// 503 (unavailable dependency).
func TestGatewayMiddlewareSmoke(t *testing.T) {
	deadConn, err := grpc.NewClient("localhost:1", grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)

	conns := &handler.ServiceConnections{Cart: deadConn}
	e := bootGateway(t, conns)
	token := signToken(t, "42")

	// Health must be reachable without a token (JWT whitelist).
	rec := perform(e, http.MethodGet, "/health", nil, "")
	require.Equal(t, http.StatusOK, rec.Code, "health must not require auth")

	// Protected route without a token → 401.
	rec = perform(e, http.MethodGet, "/api/cart-query", nil, "")
	require.Equal(t, http.StatusUnauthorized, rec.Code)

	// Unknown route with a valid token → 404.
	rec = perform(e, http.MethodGet, "/api/does-not-exist", nil, token)
	require.Equal(t, http.StatusNotFound, rec.Code)

	// Invalid JSON body with a valid token → 400 (before the gRPC call).
	rec = perform(e, http.MethodPost, "/api/cart-command/create", strings.NewReader("{not-json"), token)
	require.Equal(t, http.StatusBadRequest, rec.Code)

	// Valid token + unavailable downstream → 503.
	rec = perform(e, http.MethodPost, "/api/cart-command/create", strings.NewReader(`{"product_id":1,"quantity":1}`), token)
	require.Equal(t, http.StatusServiceUnavailable, rec.Code, "unavailable gRPC dependency must map to 503")
}
