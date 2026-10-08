package middlewares

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	apicache "github.com/MamangRust/microservice-ecommerce-grpc-apigateway/cache"
	"github.com/MamangRust/microservice-ecommerce-grpc-apigateway/httpx"
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	pbuserrole "github.com/MamangRust/microservice-ecommerce-grpc-pb/user_role"
	"go.uber.org/zap"
)

// RoleValidatorGRPC validates a user's roles via the user_role service gRPC
// (FindByUserId) and stores the role names in the request context.
//
// Ini menggantikan RoleValidator berbasis Kafka (request-response ke topic
// "request-role"/"response-role") yang tidak berfungsi di stack lokal sehingga
// semua route admin memakai middleware ini timeout 408. Dengan gRPC langsung,
// role diverifikasi secara sinkron dan RequireRoles dapat memutuskan 403.
func RoleValidatorGRPC(client pbuserrole.UserRoleServiceClient, logger logger.LoggerInterface, cache apicache.RoleCache) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userIDVal := httpx.Get(r, "user_id")
			if userIDVal == nil {
				httpx.WriteHTTPError(w, httpx.NewHTTPError(http.StatusUnauthorized, "User ID not found in context"))
				return
			}

			userID, err := extractUserIDGeneric(userIDVal)
			if err != nil {
				logger.Error("Invalid User ID format", zap.Any("value", userIDVal), zap.Error(err))
				httpx.WriteHTTPError(w, httpx.NewHTTPError(http.StatusUnauthorized, "Invalid User ID format"))
				return
			}

			ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
			defer cancel()

			if roles, found := cache.GetRoleCache(ctx, strconv.Itoa(userID)); found {
				r = r.WithContext(httpx.SetValue(r.Context(), "role_names", roles))
				next.ServeHTTP(w, r)
				return
			}

			res, err := client.FindByUserId(ctx, &pbuserrole.FindByIdUserRoleRequest{UserId: int32(userID)})
			if err != nil {
				logger.Error("Role validation via gRPC failed",
					zap.Int("user_id", userID), zap.Error(err))
				httpx.WriteHTTPError(w, httpx.NewHTTPError(http.StatusInternalServerError, "Role validation failed"))
				return
			}

			roles := make([]string, 0, len(res.GetData()))
			for _, rg := range res.GetData() {
				roles = append(roles, rg.GetName())
			}
			if len(roles) == 0 {
				logger.Debug("Role validation failed (no roles)",
					zap.Int("user_id", userID))
				httpx.WriteHTTPError(w, httpx.NewHTTPError(http.StatusUnauthorized, "Role validation failed"))
				return
			}

			cache.SetRoleCache(ctx, strconv.Itoa(userID), roles)
			r = r.WithContext(httpx.SetValue(r.Context(), "role_names", roles))
			next.ServeHTTP(w, r)
		})
	}
}

func extractUserIDGeneric(userIDVal interface{}) (int, error) {
	switch val := userIDVal.(type) {
	case float64:
		return int(val), nil
	case int:
		return val, nil
	case int32:
		return int(val), nil
	case string:
		return strconv.Atoi(val)
	default:
		return 0, fmt.Errorf("unknown user ID type %T", userIDVal)
	}
}
