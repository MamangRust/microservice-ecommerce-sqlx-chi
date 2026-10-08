package userhandler

import (
	"net/http"
	"strconv"

	"github.com/MamangRust/microservice-ecommerce-grpc-apigateway/apierror"
	user_cache "github.com/MamangRust/microservice-ecommerce-grpc-apigateway/cache/user"
	"github.com/MamangRust/microservice-ecommerce-grpc-apigateway/httpx"
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
	"github.com/MamangRust/microservice-ecommerce-shared/errors"
	apimapper "github.com/MamangRust/microservice-ecommerce-shared/mapper/user"
	pbuser "github.com/MamangRust/microservice-ecommerce-grpc-pb/user"
	"github.com/go-chi/chi/v5"
)

type userQueryHandleApi struct {
	client pbuser.UserQueryServiceClient

	logger logger.LoggerInterface

	mapper apimapper.UserQueryResponseMapper

	cache user_cache.UserMencache

	apiHandler apierror.ApiHandler
}

type userQueryHandleDeps struct {
	client pbuser.UserQueryServiceClient

	router chi.Router

	logger logger.LoggerInterface

	mapper apimapper.UserQueryResponseMapper

	cache user_cache.UserMencache

	apiHandler apierror.ApiHandler
}

func NewUserQueryHandleApi(params *userQueryHandleDeps) *userQueryHandleApi {

	userQueryHandleApi := &userQueryHandleApi{
		client:     params.client,
		logger:     params.logger,
		mapper:     params.mapper,
		cache:      params.cache,
		apiHandler: params.apiHandler,
	}

	params.router.Route("/api/user-query", func(routerUser chi.Router) {

		routerUser.Get("/", params.apiHandler.Handle("find-all-users", userQueryHandleApi.FindAllUser))
		routerUser.Get("/{id}", params.apiHandler.Handle("find-user-by-id", userQueryHandleApi.FindById))
		routerUser.Get("/active", params.apiHandler.Handle("find-active-users", userQueryHandleApi.FindByActive))
		routerUser.Get("/trashed", params.apiHandler.Handle("find-trashed-users", userQueryHandleApi.FindByTrashed))

	})
	return userQueryHandleApi
}

// @Security Bearer
// @Summary Find all users
// @Tags User Query
// @Description Retrieve a list of all users
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Number of items per page" default(10)
// @Param search query string false "Search query"
// @Success 200 {object} response.ApiResponsePaginationUser "List of users"
// @Failure 500 {object} response.ErrorResponse "Failed to retrieve user data"
// @Router /api/user-query [get]
func (h *userQueryHandleApi) FindAllUser(w http.ResponseWriter, r *http.Request) error {
	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil || page <= 0 {
		page = 1
	}

	pageSize, err := strconv.Atoi(r.URL.Query().Get("page_size"))
	if err != nil || pageSize <= 0 {
		pageSize = 10
	}

	search := r.URL.Query().Get("search")

	ctx := r.Context()

	req := &requests.FindAllUsers{
		Page:     page,
		PageSize: pageSize,
		Search:   search,
	}

	cachedData, found := h.cache.GetCachedUsersCache(ctx, req)
	if found {
		return httpx.JSON(w, http.StatusOK, cachedData)
	}

	grpcReq := &pbuser.FindAllUserRequest{
		Page:     int32(page),
		PageSize: int32(pageSize),
		Search:   search,
	}

	res, err := h.client.FindAll(ctx, grpcReq)
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	apiResponse := h.mapper.ToApiResponsePaginationUser(res)

	h.cache.SetCachedUsersCache(ctx, req, apiResponse)

	return httpx.JSON(w, http.StatusOK, apiResponse)
}

// @Security Bearer
// @Summary Find user by ID
// @Tags User Query
// @Description Retrieve a user by ID
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} response.ApiResponseUser "User data"
// @Failure 400 {object} response.ErrorResponse "Invalid user ID"
// @Failure 500 {object} response.ErrorResponse "Failed to retrieve user data"
// @Router /api/user-query/{id} [get]
func (h *userQueryHandleApi) FindById(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return errors.NewBadRequestError("id is required")
	}

	ctx := r.Context()

	cachedData, found := h.cache.GetCachedUserCache(ctx, id)
	if found {
		return httpx.JSON(w, http.StatusOK, cachedData)
	}

	req := &pbuser.FindByIdUserRequest{
		Id: int32(id),
	}

	user, err := h.client.FindById(ctx, req)
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	apiResponse := h.mapper.ToApiResponseUser(user)

	h.cache.SetCachedUserCache(ctx, apiResponse)

	return httpx.JSON(w, http.StatusOK, apiResponse)
}

// @Security Bearer
// @Summary Retrieve active users
// @Tags User Query
// @Description Retrieve a list of active users
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Number of items per page" default(10)
// @Param search query string false "Search query"
// @Success 200 {object} response.ApiResponsePaginationUserDeleteAt "List of active users"
// @Failure 500 {object} response.ErrorResponse "Failed to retrieve user data"
// @Router /api/user-query/active [get]
func (h *userQueryHandleApi) FindByActive(w http.ResponseWriter, r *http.Request) error {
	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil || page <= 0 {
		page = 1
	}

	pageSize, err := strconv.Atoi(r.URL.Query().Get("page_size"))
	if err != nil || pageSize <= 0 {
		pageSize = 10
	}

	search := r.URL.Query().Get("search")

	ctx := r.Context()

	req := &requests.FindAllUsers{
		Page:     page,
		PageSize: pageSize,
		Search:   search,
	}

	cachedData, found := h.cache.GetCachedUserActiveCache(ctx, req)
	if found {
		return httpx.JSON(w, http.StatusOK, cachedData)
	}

	grpcReq := &pbuser.FindAllUserRequest{
		Page:     int32(page),
		PageSize: int32(pageSize),
		Search:   search,
	}

	res, err := h.client.FindByActive(ctx, grpcReq)
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	apiResponse := h.mapper.ToApiResponsePaginationUserDeleteAt(res)

	h.cache.SetCachedUserActiveCache(ctx, req, apiResponse)

	return httpx.JSON(w, http.StatusOK, apiResponse)
}

// @Security Bearer
// @Summary Retrieve trashed users
// @Tags User Query
// @Description Retrieve a list of trashed user records
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Number of items per page" default(10)
// @Param search query string false "Search query"
// @Success 200 {object} response.ApiResponsePaginationUserDeleteAt "List of trashed user data"
// @Failure 500 {object} response.ErrorResponse "Failed to retrieve user data"
// @Router /api/user-query/trashed [get]
func (h *userQueryHandleApi) FindByTrashed(w http.ResponseWriter, r *http.Request) error {
	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil || page <= 0 {
		page = 1
	}

	pageSize, err := strconv.Atoi(r.URL.Query().Get("page_size"))
	if err != nil || pageSize <= 0 {
		pageSize = 10
	}

	search := r.URL.Query().Get("search")

	ctx := r.Context()

	req := &requests.FindAllUsers{
		Page:     page,
		PageSize: pageSize,
		Search:   search,
	}

	cachedData, found := h.cache.GetCachedUserTrashedCache(ctx, req)
	if found {
		return httpx.JSON(w, http.StatusOK, cachedData)
	}

	grpcReq := &pbuser.FindAllUserRequest{
		Page:     int32(page),
		PageSize: int32(pageSize),
		Search:   search,
	}

	res, err := h.client.FindByTrashed(ctx, grpcReq)
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	apiResponse := h.mapper.ToApiResponsePaginationUserDeleteAt(res)

	h.cache.SetCachedUserTrashedCache(ctx, req, apiResponse)

	return httpx.JSON(w, http.StatusOK, apiResponse)
}
