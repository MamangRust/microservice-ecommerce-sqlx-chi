package middlewares

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/MamangRust/microservice-ecommerce-grpc-apigateway/httpx"
	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/viper"
)

var whiteListPaths = []string{
	"/api/auth/login",
	"/api/auth/register",
	"/api/auth/hello",
	"/api/auth/refresh-token",
	"/api/auth/verify-code",
	"/health",
	"/healthz",
	"/ready",
	"/docs/",
	"/docs",
	"/swagger",
	"/metrics",
}

func extractUserIDFromClaims(claims jwt.MapClaims) (int, bool) {
	subVal, exists := claims["sub"]
	if !exists {
		return 0, false
	}

	switch v := subVal.(type) {
	case string:
		if v == "" {
			return 0, false
		}
		id, err := strconv.Atoi(v)
		if err != nil {
			return 0, false
		}
		return id, true
	case float64:
		return int(v), true
	case int:
		return v, true
	default:
		return 0, false
	}
}

// JWTAuth validates the Bearer access token and stores the numeric user ID in
// the request context under "user_id", matching the request-scoped value that
// echojwt previously set on echo.Context.
func JWTAuth() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if skipAuth(r) {
				next.ServeHTTP(w, r)
				return
			}

			tokenString, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
			if !ok || tokenString == "" {
				unauthorized(w)
				return
			}

			token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
				}
				return []byte(viper.GetString("SECRET_KEY")), nil
			})
			if err != nil || !token.Valid {
				unauthorized(w)
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				unauthorized(w)
				return
			}
			if id, ok := extractUserIDFromClaims(claims); ok && id > 0 {
				r = r.WithContext(httpx.SetValue(r.Context(), "user_id", id))
			}

			next.ServeHTTP(w, r)
		})
	}
}

func skipAuth(r *http.Request) bool {
	path := r.URL.Path

	for _, p := range whiteListPaths {
		if path == p ||
			strings.HasPrefix(path, "/swagger") || strings.HasPrefix(path, "/api/auth/verify-code") {
			return true
		}
	}

	return false
}

func unauthorized(w http.ResponseWriter) {
	_ = httpx.JSON(w, http.StatusUnauthorized, map[string]string{"message": "Unauthorized"})
}
