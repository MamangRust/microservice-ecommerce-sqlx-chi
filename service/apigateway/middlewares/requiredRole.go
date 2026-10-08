package middlewares

import (
	"net/http"

	"github.com/MamangRust/microservice-ecommerce-grpc-apigateway/httpx"
)

func RequireRoles(allowedRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			raw, ok := httpx.Get(r, "role_names").([]string)
			if !ok {
				httpx.WriteHTTPError(w, httpx.NewHTTPError(http.StatusForbidden, "Roles not found"))
				return
			}

			for _, role := range raw {
				for _, allowed := range allowedRoles {
					if role == allowed {
						next.ServeHTTP(w, r)
						return
					}
				}
			}

			httpx.WriteHTTPError(w, httpx.NewHTTPError(http.StatusForbidden, "Role not permitted"))
		})
	}
}
