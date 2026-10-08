package userhandler

import (
	"fmt"
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
	"github.com/go-playground/validator/v10"
	"google.golang.org/protobuf/types/known/emptypb"
)

type userCommandHandleApi struct {
	client pbuser.UserCommandServiceClient

	logger logger.LoggerInterface

	mapper apimapper.UserCommandResponseMapper

	cache user_cache.UserMencache

	apiHandler apierror.ApiHandler
}

type userCommandHandleDeps struct {
	client pbuser.UserCommandServiceClient

	router chi.Router

	logger logger.LoggerInterface

	mapper apimapper.UserCommandResponseMapper

	cache user_cache.UserMencache

	apiHandler apierror.ApiHandler
}

func NewUserCommandHandleApi(params *userCommandHandleDeps) *userCommandHandleApi {

	userCommandHandleApi := &userCommandHandleApi{
		client:     params.client,
		logger:     params.logger,
		mapper:     params.mapper,
		cache:      params.cache,
		apiHandler: params.apiHandler,
	}

	params.router.Route("/api/user-command", func(routerUser chi.Router) {

		routerUser.Post("/create", params.apiHandler.Handle("create-user", userCommandHandleApi.Create))
		routerUser.Post("/update/{id}", params.apiHandler.Handle("update-user", userCommandHandleApi.Update))

		routerUser.Post("/trashed/{id}", params.apiHandler.Handle("trash-user", userCommandHandleApi.TrashedUser))
		routerUser.Post("/restore/{id}", params.apiHandler.Handle("restore-user", userCommandHandleApi.RestoreUser))
		routerUser.Delete("/permanent/{id}", params.apiHandler.Handle("delete-user-permanent", userCommandHandleApi.DeleteUserPermanent))

		routerUser.Post("/restore/all", params.apiHandler.Handle("restore-all-users", userCommandHandleApi.RestoreAllUser))
		routerUser.Post("/permanent/all", params.apiHandler.Handle("delete-all-users-permanent", userCommandHandleApi.DeleteAllUserPermanent))

	})
	return userCommandHandleApi
}

// @Security Bearer
// Create handles the creation of a new user.
// @Summary Create a new user
// @Tags User Command
// @Description Create a new user with the provided details
// @Accept json
// @Produce json
// @Param request body requests.CreateUserRequest true "Create user request"
// @Success 200 {object} response.ApiResponseUser "Successfully created user"
// @Failure 400 {object} response.ErrorResponse "Invalid request body or validation error"
// @Failure 500 {object} response.ErrorResponse "Failed to create user"
// @Router /api/user-command/create [post]
func (h *userCommandHandleApi) Create(w http.ResponseWriter, r *http.Request) error {
	var body requests.CreateUserRequest

	if err := httpx.Bind(r, &body); err != nil {
		return errors.NewBadRequestError("Invalid request format").WithInternal(err)
	}

	if err := body.Validate(); err != nil {
		validations := h.parseValidationErrors(err)
		return errors.NewValidationError(validations)
	}

	ctx := r.Context()

	req := &pbuser.CreateUserRequest{
		Firstname:       body.FirstName,
		Lastname:        body.LastName,
		Email:           body.Email,
		Password:        body.Password,
		ConfirmPassword: body.ConfirmPassword,
	}

	res, err := h.client.Create(ctx, req)

	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapper.ToApiResponseUser(res)

	h.cache.SetCachedUserCache(ctx, so)

	return httpx.JSON(w, http.StatusOK, so)
}

// @Security Bearer
// Update handles the update of an existing user record.
// @Summary Update an existing user
// @Tags User Command
// @Description Update an existing user record with the provided details
// @Accept json
// @Produce json
// @Param UpdateUserRequest body requests.UpdateUserRequest true "Update user request"
// @Success 200 {object} response.ApiResponseUser "Successfully updated user"
// @Failure 400 {object} response.ErrorResponse "Invalid request body or validation error"
// @Failure 500 {object} response.ErrorResponse "Failed to update user"
// @Router /api/user-command/update/{id} [post]
func (h *userCommandHandleApi) Update(w http.ResponseWriter, r *http.Request) error {
	id := chi.URLParam(r, "id")

	idInt, err := strconv.Atoi(id)

	if err != nil {
		return errors.NewBadRequestError("id is required")
	}

	var body requests.UpdateUserRequest

	if err := httpx.Bind(r, &body); err != nil {
		return errors.NewBadRequestError("Invalid request format").WithInternal(err)
	}

	if err := body.Validate(); err != nil {
		validations := h.parseValidationErrors(err)
		return errors.NewValidationError(validations)
	}

	ctx := r.Context()

	req := &pbuser.UpdateUserRequest{
		Id:              int32(idInt),
		Firstname:       body.FirstName,
		Lastname:        body.LastName,
		Email:           body.Email,
		Password:        body.Password,
		ConfirmPassword: body.ConfirmPassword,
	}

	res, err := h.client.Update(ctx, req)

	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapper.ToApiResponseUser(res)

	h.cache.DeleteUserCache(ctx, idInt)
	h.cache.SetCachedUserCache(ctx, so)

	return httpx.JSON(w, http.StatusOK, so)
}

// @Security Bearer
// @Summary Retrieve a trashed user
// @Tags User Command
// @Description Retrieve a trashed user record by its ID.
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} response.ApiResponseUserDeleteAt "Successfully retrieved trashed user"
// @Failure 400 {object} response.ErrorResponse "Invalid request body or validation error"
// @Failure 500 {object} response.ErrorResponse "Failed to retrieve trashed user"
// @Router /api/user-command/trashed/{id} [get]
func (h *userCommandHandleApi) TrashedUser(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))

	if err != nil {
		return errors.NewBadRequestError("id is required")
	}

	ctx := r.Context()

	req := &pbuser.FindByIdUserRequest{
		Id: int32(id),
	}

	user, err := h.client.TrashedUser(ctx, req)

	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapper.ToApiResponseUserDeleteAt(user)

	h.cache.DeleteUserCache(ctx, id)

	return httpx.JSON(w, http.StatusOK, so)
}

// @Security Bearer
// @Summary Restore a trashed user
// @Tags User Command
// @Description Restore a trashed user record by its ID.
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} response.ApiResponseUserDeleteAt "Successfully restored user"
// @Failure 400 {object} response.ErrorResponse "Invalid user ID"
// @Failure 500 {object} response.ErrorResponse "Failed to restore user"
// @Router /api/user-command/restore/{id} [post]
func (h *userCommandHandleApi) RestoreUser(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))

	if err != nil {
		return errors.NewBadRequestError("id is required")
	}

	ctx := r.Context()

	req := &pbuser.FindByIdUserRequest{
		Id: int32(id),
	}

	user, err := h.client.RestoreUser(ctx, req)

	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapper.ToApiResponseUserDeleteAt(user)

	h.cache.DeleteUserCache(ctx, id)

	return httpx.JSON(w, http.StatusOK, so)
}

// @Security Bearer
// @Summary Permanently delete a user
// @Tags User Command
// @Description Permanently delete a user record by its ID.
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} response.ApiResponseUserDelete "Successfully deleted user record permanently"
// @Failure 400 {object} response.ErrorResponse "Bad Request: Invalid ID"
// @Failure 500 {object} response.ErrorResponse "Failed to delete user"
// @Router /api/user-command/permanent/{id} [delete]
func (h *userCommandHandleApi) DeleteUserPermanent(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))

	if err != nil {
		return errors.NewBadRequestError("id is required")
	}

	ctx := r.Context()

	req := &pbuser.FindByIdUserRequest{
		Id: int32(id),
	}

	user, err := h.client.DeleteUserPermanent(ctx, req)

	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapper.ToApiResponseUserDelete(user)

	h.cache.DeleteUserCache(ctx, id)

	return httpx.JSON(w, http.StatusOK, so)
}

// @Security Bearer
// @Summary Restore all trashed users
// @Tags User Command
// @Description Restore all trashed user records.
// @Accept json
// @Produce json
// @Success 200 {object} response.ApiResponseUserAll "Successfully restored all users"
// @Failure 500 {object} response.ErrorResponse "Failed to restore users"
// @Router /api/user-command/restore/all [post]
func (h *userCommandHandleApi) RestoreAllUser(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	res, err := h.client.RestoreAllUser(ctx, &emptypb.Empty{})

	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapper.ToApiResponseUserAll(res)

	h.logger.Debug("Successfully restored all user")

	return httpx.JSON(w, http.StatusOK, so)
}

// @Security Bearer
// @Summary Permanently delete all trashed users
// @Tags User Command
// @Description Permanently delete all trashed user records.
// @Accept json
// @Produce json
// @Success 200 {object} response.ApiResponseUserAll "Successfully deleted all users permanently"
// @Failure 500 {object} response.ErrorResponse "Failed to delete users"
// @Router /api/user-command/permanent/all [post]
func (h *userCommandHandleApi) DeleteAllUserPermanent(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	res, err := h.client.DeleteAllUserPermanent(ctx, &emptypb.Empty{})

	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapper.ToApiResponseUserAll(res)

	h.logger.Debug("Successfully deleted all user permanently")

	return httpx.JSON(w, http.StatusOK, so)
}

func (h *userCommandHandleApi) parseValidationErrors(err error) []errors.ValidationError {
	var validationErrs []errors.ValidationError

	if ve, ok := err.(validator.ValidationErrors); ok {
		for _, fe := range ve {
			validationErrs = append(validationErrs, errors.ValidationError{
				Field:   fe.Field(),
				Message: h.getValidationMessage(fe),
			})
		}
		return validationErrs
	}

	return []errors.ValidationError{
		{
			Field:   "general",
			Message: err.Error(),
		},
	}
}

func (h *userCommandHandleApi) getValidationMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "This field is required"
	case "email":
		return "Invalid email format"
	case "min":
		return fmt.Sprintf("Must be at least %s", fe.Param())
	case "max":
		return fmt.Sprintf("Must be at most %s", fe.Param())
	case "gte":
		return fmt.Sprintf("Must be greater than or equal to %s", fe.Param())
	case "lte":
		return fmt.Sprintf("Must be less than or equal to %s", fe.Param())
	case "oneof":
		return fmt.Sprintf("Must be one of: %s", fe.Param())
	default:
		return fmt.Sprintf("Validation failed on '%s' tag", fe.Tag())
	}
}
