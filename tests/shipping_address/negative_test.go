package shipping_address_test

import (
	"context"
	"github.com/MamangRust/microservice-ecommerce-shared/errors"
	"net/http"
	"net/http/httptest"

	"github.com/MamangRust/microservice-ecommerce-grpc-pb/shipping_address"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// gapi: non-existent shipping address must map to codes.NotFound (404), not Internal.
func (s *ShippingAddressGapiTestSuite) TestShippingAddressGapiNotFound() {
	ctx := context.Background()
	_, err := s.queryClient.FindById(ctx, &pb_shipping_address.FindByIdShippingRequest{Id: 999999})
	s.Require().Error(err)
	st, ok := status.FromError(err)
	s.Require().True(ok, "expected a gRPC status error")
	s.Equal(codes.NotFound, st.Code(), "non-existent shipping address must be NotFound, got %v: %s", st.Code(), st.Message())
}

// api: non-existent shipping address must map to 404, invalid path ID to 400.
func (s *ShippingAddressApiTestSuite) TestShippingAddressApiNotFound() {

	req := httptest.NewRequest(http.MethodGet, "/api/shipping-address-query/999999", nil)
	rec := httptest.NewRecorder()
	s.server.ServeHTTP(rec, req)
	s.Equal(http.StatusNotFound, rec.Code, "non-existent shipping address must be 404, got %d: %s", rec.Code, rec.Body.String())
}

func (s *ShippingAddressApiTestSuite) TestShippingAddressApiInvalidID() {

	req := httptest.NewRequest(http.MethodGet, "/api/shipping-address-query/abc", nil)
	rec := httptest.NewRecorder()
	s.server.ServeHTTP(rec, req)
	s.Equal(http.StatusBadRequest, rec.Code, "invalid shipping address ID must be 400, got %d: %s", rec.Code, rec.Body.String())
}

// repository: FindByID on a non-existent ID must return a typed not-found error.
func (s *ShippingAddressRepositoryTestSuite) TestShippingAddressFindByIDNotFound() {
	ctx := context.Background()
	_, err := s.repo.ShippingAddressQuery.FindByID(ctx, 999999)
	s.Require().Error(err)
	var appErr *errors.AppError
	s.Require().ErrorAs(err, &appErr)
	s.Equal(errors.ErrorTypeNotFound, appErr.Type, "expected not-found error type, got %s: %v", appErr.Type, err)
}
