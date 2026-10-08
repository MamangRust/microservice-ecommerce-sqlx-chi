package authhandler

import (
	"net/http"
	"strconv"

	"fmt"

	"github.com/MamangRust/microservice-ecommerce-grpc-apigateway/apierror"
	auth_cache "github.com/MamangRust/microservice-ecommerce-grpc-apigateway/cache/auth"
	"github.com/MamangRust/microservice-ecommerce-grpc-apigateway/httpx"
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
	sharedErrors "github.com/MamangRust/microservice-ecommerce-shared/errors"
	authapimapper "github.com/MamangRust/microservice-ecommerce-shared/mapper/auth"
	pbauth "github.com/MamangRust/microservice-ecommerce-grpc-pb/auth"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type authHandleParams struct {
	client     pbauth.AuthServiceClient
	router     chi.Router
	cache      auth_cache.AuthMencache
	logger     logger.LoggerInterface
	mapper     authapimapper.AuthResponseMapper
	apiHandler apierror.ApiHandler
}

type authHandleApi struct {
	client        pbauth.AuthServiceClient
	logger        logger.LoggerInterface
	queryMapper   authapimapper.AuthQueryResponseMapper
	commandMapper authapimapper.AuthCommandResponseMapper
	apiHandler    apierror.ApiHandler
	cache         auth_cache.AuthMencache
}

type DepsAuth struct {
	Client     *grpc.ClientConn
	Router     chi.Router
	Logger     logger.LoggerInterface
	Cache      auth_cache.AuthMencache
	ApiHandler apierror.ApiHandler
}

func RegisterAuthHandler(deps *DepsAuth) {
	mapper := authapimapper.NewAuthResponseMapper()

	NewHandlerAuth(&authHandleParams{
		client:     pbauth.NewAuthServiceClient(deps.Client),
		router:     deps.Router,
		cache:      deps.Cache,
		logger:     deps.Logger,
		mapper:     mapper,
		apiHandler: deps.ApiHandler,
	})
}

func NewHandlerAuth(params *authHandleParams) *authHandleApi {
	authHandler := &authHandleApi{
		client:        params.client,
		logger:        params.logger,
		queryMapper:   params.mapper.QueryMapper(),
		commandMapper: params.mapper.CommandMapper(),
		apiHandler:    params.apiHandler,
		cache:         params.cache,
	}
	params.router.Route("/api/auth", func(routerAuth chi.Router) {

		routerAuth.Get("/hello", httpx.Handler(authHandler.HandleHello))
		routerAuth.Post("/register", params.apiHandler.Handle("register", authHandler.Register))
		routerAuth.Post("/login", params.apiHandler.Handle("login", authHandler.Login))
		routerAuth.Post("/refresh-token", params.apiHandler.Handle("refresh-token", authHandler.RefreshToken))
		routerAuth.Get("/me", params.apiHandler.Handle("GetMe", authHandler.GetMe))

	})
	return authHandler
}

// @Summary Auth endpoint liveness check
// @Tags Auth
// @Description Simple liveness probe for the auth endpoint
// @Produce plain
// @Success 200 {string} string "Hello"
// @Router /api/auth/hello [get]
func (h *authHandleApi) HandleHello(w http.ResponseWriter, r *http.Request) error {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("Hello"))
	return nil
}

// @Summary Register a new user
// @Tags Auth
// @Description Create a new user account
// @Accept json
// @Produce json
// @Param request body requests.CreateUserRequest true "Registration details"
// @Success 201 {object} response.ApiResponseRegister "Successfully registered"
// @Failure 400 {object} errors.ErrorResponse "Validation error"
// @Failure 500 {object} errors.ErrorResponse "Internal server error"
// @Router /api/auth/register [post]
func (h *authHandleApi) Register(w http.ResponseWriter, r *http.Request) error {
	var body requests.CreateUserRequest

	if err := httpx.Bind(r, &body); err != nil {
		return sharedErrors.NewBadRequestError("Invalid request format").WithInternal(err)
	}

	if err := body.Validate(); err != nil {
		validations := h.parseValidationErrors(err)
		return sharedErrors.NewValidationError(validations)
	}

	data := &pbauth.RegisterRequest{
		Firstname:       body.FirstName,
		Lastname:        body.LastName,
		Email:           body.Email,
		Password:        body.Password,
		ConfirmPassword: body.ConfirmPassword,
	}

	res, err := h.client.RegisterUser(r.Context(), data)
	if err != nil {
		return sharedErrors.ParseGrpcError(err)
	}

	return httpx.JSON(w, http.StatusCreated, h.commandMapper.ToResponseRegister(res))
}

// @Summary Login user
// @Tags Auth
// @Description Authenticate user and return tokens
// @Accept json
// @Produce json
// @Param request body requests.AuthRequest true "Login credentials"
// @Success 200 {object} response.ApiResponseLogin "Successfully logged in"
// @Failure 401 {object} errors.ErrorResponse "Unauthorized"
// @Failure 500 {object} errors.ErrorResponse "Internal server error"
// @Router /api/auth/login [post]
func (h *authHandleApi) Login(w http.ResponseWriter, r *http.Request) error {
	var body requests.AuthRequest

	if err := httpx.Bind(r, &body); err != nil {
		return sharedErrors.NewBadRequestError("Invalid request format").WithInternal(err)
	}

	if err := body.Validate(); err != nil {
		validations := h.parseValidationErrors(err)
		return sharedErrors.NewValidationError(validations)
	}

	ctx := r.Context()
	cachedResponse, found := h.cache.GetCachedLogin(ctx, body.Email)
	if found {
		h.logger.Debug("Returning login response from cache", zap.String("email", body.Email))
		return httpx.JSON(w, http.StatusOK, cachedResponse)
	}

	res, err := h.client.LoginUser(ctx, &pbauth.LoginRequest{
		Email:    body.Email,
		Password: body.Password,
	})

	if err != nil {
		return sharedErrors.ParseGrpcError(err)
	}

	mappedResponse := h.commandMapper.ToResponseLogin(res)
	h.cache.SetCachedLogin(ctx, body.Email, mappedResponse)

	return httpx.JSON(w, http.StatusOK, mappedResponse)
}

// @Summary Refresh token
// @Tags Auth
// @Description refresh token
// @Accept json
// @Produce json
// @Param request body requests.RefreshTokenRequest true "Refresh token details"
// @Success 200 {object} response.ApiResponseRefreshToken "Successfully refreshed token"
// @Failure 401 {object} errors.ErrorResponse "Unauthorized"
// @Failure 500 {object} errors.ErrorResponse "Internal server error"
// @Router /api/auth/refresh-token [post]
func (h *authHandleApi) RefreshToken(w http.ResponseWriter, r *http.Request) error {
	var body requests.RefreshTokenRequest

	if err := httpx.Bind(r, &body); err != nil {
		return sharedErrors.NewBadRequestError("Invalid request format").WithInternal(err)
	}

	if err := body.Validate(); err != nil {
		validations := h.parseValidationErrors(err)
		return sharedErrors.NewValidationError(validations)
	}

	ctx := r.Context()
	cachedResponse, found := h.cache.GetRefreshToken(ctx, body.RefreshToken)
	if found {
		h.logger.Debug("Returning refresh token response from cache")
		return httpx.JSON(w, http.StatusOK, cachedResponse)
	}

	res, err := h.client.RefreshToken(ctx, &pbauth.RefreshTokenRequest{
		RefreshToken: body.RefreshToken,
	})
	if err != nil {
		return sharedErrors.ParseGrpcError(err)
	}

	mappedResponse := h.commandMapper.ToResponseRefreshToken(res)
	h.cache.SetRefreshToken(ctx, body.RefreshToken, mappedResponse)

	return httpx.JSON(w, http.StatusOK, mappedResponse)
}

// @Security Bearer
// @Summary Get current user info
// @Tags Auth
// @Description Retrieve current authenticated user details
// @Accept json
// @Produce json
// @Success 200 {object} response.ApiResponseGetMe "User info"
// @Failure 401 {object} errors.ErrorResponse "Unauthorized"
// @Failure 500 {object} errors.ErrorResponse "Internal server error"
// @Router /api/auth/me [get]
func (h *authHandleApi) GetMe(w http.ResponseWriter, r *http.Request) error {
	// The JWT middleware stores the authenticated user id as an int under
	// "user_id" (see middlewares/auth.go). Reading any other key here would
	// make /me always fail authentication.
	userID, ok := httpx.Get(r, "user_id").(int)
	if !ok || userID <= 0 {
		return sharedErrors.NewBadRequestError("user not authenticated")
	}

	ctx := r.Context()
	if cached, found := h.cache.GetCachedUserInfo(ctx, strconv.Itoa(userID)); found {
		return httpx.JSON(w, http.StatusOK, cached)
	}

	res, err := h.client.GetMe(ctx, &pbauth.GetMeRequest{UserId: int32(userID)})
	if err != nil {
		return sharedErrors.ParseGrpcError(err)
	}

	response := h.queryMapper.ToResponseGetMe(res)
	h.cache.SetCachedUserInfo(ctx, strconv.Itoa(userID), response)

	return httpx.JSON(w, http.StatusOK, response)
}

func (h *authHandleApi) parseValidationErrors(err error) []sharedErrors.ValidationError {
	var validationErrs []sharedErrors.ValidationError

	if ve, ok := err.(validator.ValidationErrors); ok {
		for _, fe := range ve {
			validationErrs = append(validationErrs, sharedErrors.ValidationError{
				Field:   fe.Field(),
				Message: h.getValidationMessage(fe),
			})
		}
		return validationErrs
	}

	return []sharedErrors.ValidationError{
		{
			Field:   "general",
			Message: err.Error(),
		},
	}
}

func (h *authHandleApi) getValidationMessage(fe validator.FieldError) string {
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
