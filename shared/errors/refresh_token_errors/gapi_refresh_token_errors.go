package refreshtoken_errors

import (
	"github.com/MamangRust/microservice-ecommerce-shared/errors"

	"net/http"
)

var ErrGrpcRefreshToken = errors.NewGrpcError("refresh token failed", http.StatusUnauthorized)
