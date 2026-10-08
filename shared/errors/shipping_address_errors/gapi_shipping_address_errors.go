package shippingaddress_errors

import (
	"github.com/MamangRust/microservice-ecommerce-shared/errors"

	"net/http"
)

var (
	ErrGrpcInvalidID = errors.NewGrpcError("invalid ID", http.StatusBadRequest)
)
