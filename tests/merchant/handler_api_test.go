package merchant_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"net/http"
	"net/http/httptest"
	"testing"

	apierror "github.com/MamangRust/microservice-ecommerce-grpc-apigateway/apierror"
	merchanthandler "github.com/MamangRust/microservice-ecommerce-grpc-apigateway/handler/merchant"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
	tests "github.com/MamangRust/microservice-ecommerce-test"
	"github.com/stretchr/testify/suite"
)

type MerchantApiTestSuite struct {
	tests.BaseTestSuite
	server     chi.Router
	merchantID int
	userID     int
}

func (s *MerchantApiTestSuite) SetupSuite() {
	s.BaseTestSuite.SetupSuite()
	s.SetupUserService()
	s.SetupMerchantService()

	// Seed user
	s.userID = s.SeedUser(context.Background())

	s.server = chi.NewRouter()
	s.server.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r = r.WithContext(context.WithValue(r.Context(), "user_id", s.userID))
			next.ServeHTTP(w, r)
		})
	})

	apiHandler := apierror.NewApiHandler(s.Obs, s.Log)

	merchanthandler.RegisterMerchantHandler(&merchanthandler.DepsMerchant{
		Client:      s.Conns["merchant"],
		Router:      s.server,
		Logger:      s.Log,
		CacheStore:  s.GetCacheStore(),
		UploadImage: nil,
		ApiHandler:  apiHandler,
	})
}

func (s *MerchantApiTestSuite) TestMerchantApiLifecycle() {
	// 1. Create
	reqBody := requests.CreateMerchantRequest{
		Name:         "Test Merchant",
		Description:  "Test Description",
		Address:      "Test Address",
		ContactEmail: "merchant@example.com",
		ContactPhone: "123456789",
		Status:       "active",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/merchant-command/create", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	s.server.ServeHTTP(rec, req)
	s.Require().Equal(http.StatusCreated, rec.Code, rec.Body.String())
	var res map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &res)
	data := res["data"].(map[string]interface{})
	s.Equal(reqBody.Name, data["name"])
	s.merchantID = int(data["id"].(float64))

	// 2. FindById
	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/merchant-query/%d", s.merchantID), nil)
	rec = httptest.NewRecorder()
	s.server.ServeHTTP(rec, req)
	s.Equal(http.StatusOK, rec.Code)
	json.Unmarshal(rec.Body.Bytes(), &res)
	data = res["data"].(map[string]interface{})
	s.Equal(float64(s.merchantID), data["id"])

	// 3. FindAll
	req = httptest.NewRequest(http.MethodGet, "/api/merchant-query", nil)
	rec = httptest.NewRecorder()
	s.server.ServeHTTP(rec, req)
	s.Equal(http.StatusOK, rec.Code)

	// 4. FindByActive
	req = httptest.NewRequest(http.MethodGet, "/api/merchant-query/active", nil)
	rec = httptest.NewRecorder()
	s.server.ServeHTTP(rec, req)
	s.Equal(http.StatusOK, rec.Code)

	// 5. Update
	updateBody := requests.UpdateMerchantRequest{
		MerchantID:   &s.merchantID,
		Name:         "Updated Merchant",
		Description:  "Updated Description",
		Address:      "Updated Address",
		ContactEmail: "updated@example.com",
		ContactPhone: "987654321",
		Status:       "active",
	}
	body, _ = json.Marshal(updateBody)
	req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/merchant-command/update/%d", s.merchantID), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	s.server.ServeHTTP(rec, req)
	s.Equal(http.StatusOK, rec.Code)
	json.Unmarshal(rec.Body.Bytes(), &res)
	data = res["data"].(map[string]interface{})
	s.Equal(updateBody.Name, data["name"])

	// 6. Trash
	req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/merchant-command/trashed/%d", s.merchantID), nil)
	rec = httptest.NewRecorder()
	s.server.ServeHTTP(rec, req)
	s.Equal(http.StatusOK, rec.Code)

	// 7. FindByTrashed
	req = httptest.NewRequest(http.MethodGet, "/api/merchant-query/trashed", nil)
	rec = httptest.NewRecorder()
	s.server.ServeHTTP(rec, req)
	s.Equal(http.StatusOK, rec.Code)

	// 8. Restore
	req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/merchant-command/restore/%d", s.merchantID), nil)
	rec = httptest.NewRecorder()
	s.server.ServeHTTP(rec, req)
	s.Equal(http.StatusOK, rec.Code)

	// 9. DeletePermanent
	req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/merchant-command/trashed/%d", s.merchantID), nil)
	rec = httptest.NewRecorder()
	s.server.ServeHTTP(rec, req)
	req = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/merchant-command/permanent/%d", s.merchantID), nil)
	rec = httptest.NewRecorder()
	s.server.ServeHTTP(rec, req)
	s.Equal(http.StatusOK, rec.Code)

	// 10. RestoreAll
	req = httptest.NewRequest(http.MethodPost, "/api/merchant-command/restore/all", nil)
	rec = httptest.NewRecorder()
	s.server.ServeHTTP(rec, req)
	s.Equal(http.StatusOK, rec.Code)

	// 11. DeleteAll
	req = httptest.NewRequest(http.MethodPost, "/api/merchant-command/permanent/all", nil)
	rec = httptest.NewRecorder()
	s.server.ServeHTTP(rec, req)
	s.Equal(http.StatusOK, rec.Code)
}

func TestMerchantApiSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(MerchantApiTestSuite))
}
