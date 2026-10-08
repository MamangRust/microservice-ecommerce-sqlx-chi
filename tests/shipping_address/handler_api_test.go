package shipping_address_test

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"net/http"
	"net/http/httptest"
	"testing"

	shippingaddresshandler "github.com/MamangRust/microservice-ecommerce-grpc-apigateway/handler/shipping_address"
	tests "github.com/MamangRust/microservice-ecommerce-test"
	"github.com/stretchr/testify/suite"
)

type ShippingAddressApiTestSuite struct {
	tests.BaseTestSuite
	server            chi.Router
	orderID           int
	shippingAddressID int
}

func (s *ShippingAddressApiTestSuite) SetupSuite() {
	s.BaseTestSuite.SetupSuite()
	s.SetupUserService()
	s.SetupCategoryService()
	s.SetupMerchantService()
	s.SetupProductService()
	s.SetupOrderItemService()
	s.SetupShippingAddressService()
	s.SetupOrderService()

	// Seed dependencies
	ctx := context.Background()
	userID := s.SeedUser(ctx)
	catID := s.SeedCategory(ctx)
	merchID := s.SeedMerchant(ctx, userID)
	prodID := s.SeedProduct(ctx, merchID, catID)
	s.orderID = s.SeedOrder(ctx, userID, merchID, prodID)

	s.server = chi.NewRouter()

	shippingaddresshandler.RegisterShippingAddressHandler(&shippingaddresshandler.DepsShippingAddress{
		Client: s.Conns["shipping-address"],
		Router: s.server,
		Logger: s.Log,
		Cache:  s.GetCacheStore(),
	})
}

func (s *ShippingAddressApiTestSuite) TestShippingAddressApiLifecycle() {
	// 1. FindAll
	req := httptest.NewRequest(http.MethodGet, "/api/shipping-address-query", nil)
	rec := httptest.NewRecorder()
	s.server.ServeHTTP(rec, req)
	s.Equal(http.StatusOK, rec.Code)

	// 2. FindByOrder
	s.Require().NotZero(s.orderID)
	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/shipping-address-query/order/%d", s.orderID), nil)
	rec = httptest.NewRecorder()
	s.server.ServeHTTP(rec, req)
	s.Equal(http.StatusOK, rec.Code, rec.Body.String())
	var res map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &res)
	data := res["data"].(map[string]interface{})
	s.shippingAddressID = int(data["id"].(float64))

	// 3. FindById
	s.Require().NotZero(s.shippingAddressID)
	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/shipping-address-query/%d", s.shippingAddressID), nil)
	rec = httptest.NewRecorder()
	s.server.ServeHTTP(rec, req)
	s.Equal(http.StatusOK, rec.Code)

	// 4. FindByActive
	req = httptest.NewRequest(http.MethodGet, "/api/shipping-address-query/active", nil)
	rec = httptest.NewRecorder()
	s.server.ServeHTTP(rec, req)
	s.Equal(http.StatusOK, rec.Code)

	// 5. Trash
	req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/shipping-address-command/trashed/%d", s.shippingAddressID), nil)
	rec = httptest.NewRecorder()
	s.server.ServeHTTP(rec, req)
	s.Equal(http.StatusOK, rec.Code)

	// 6. FindByTrashed
	req = httptest.NewRequest(http.MethodGet, "/api/shipping-address-query/trashed", nil)
	rec = httptest.NewRecorder()
	s.server.ServeHTTP(rec, req)
	s.Equal(http.StatusOK, rec.Code)

	// 7. Restore
	req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/shipping-address-command/restore/%d", s.shippingAddressID), nil)
	rec = httptest.NewRecorder()
	s.server.ServeHTTP(rec, req)
	s.Equal(http.StatusOK, rec.Code)

	// 8. DeletePermanent
	req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/shipping-address-command/trashed/%d", s.shippingAddressID), nil)
	rec = httptest.NewRecorder()
	s.server.ServeHTTP(rec, req)
	req = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/shipping-address-command/permanent/%d", s.shippingAddressID), nil)
	rec = httptest.NewRecorder()
	s.server.ServeHTTP(rec, req)
	s.Equal(http.StatusOK, rec.Code)

	// 9. RestoreAll
	req = httptest.NewRequest(http.MethodPost, "/api/shipping-address-command/restore/all", nil)
	rec = httptest.NewRecorder()
	s.server.ServeHTTP(rec, req)
	s.Equal(http.StatusOK, rec.Code)

	// 10. DeleteAll
	req = httptest.NewRequest(http.MethodPost, "/api/shipping-address-command/permanent/all", nil)
	rec = httptest.NewRecorder()
	s.server.ServeHTTP(rec, req)
	s.Equal(http.StatusOK, rec.Code)
}

func TestShippingAddressApiSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(ShippingAddressApiTestSuite))
}
