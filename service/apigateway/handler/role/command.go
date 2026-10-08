package rolehandler

import (
	"net/http"
	"strconv"

	"github.com/MamangRust/microservice-ecommerce-grpc-apigateway/apierror"
	mencache "github.com/MamangRust/microservice-ecommerce-grpc-apigateway/cache"
	role_cache "github.com/MamangRust/microservice-ecommerce-grpc-apigateway/cache/role"
	"github.com/MamangRust/microservice-ecommerce-grpc-apigateway/httpx"
	"github.com/MamangRust/microservice-ecommerce-grpc-apigateway/middlewares"
	"github.com/MamangRust/microservice-ecommerce-pkg/kafka"
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
	"github.com/MamangRust/microservice-ecommerce-shared/errors"
	apimapper "github.com/MamangRust/microservice-ecommerce-shared/mapper/role"
	pbrole "github.com/MamangRust/microservice-ecommerce-grpc-pb/role"
	pbuserrole "github.com/MamangRust/microservice-ecommerce-grpc-pb/user_role"
	"github.com/go-chi/chi/v5"
	"google.golang.org/protobuf/types/known/emptypb"
)

type roleCommandHandlerApi struct {
	kafka      *kafka.Kafka
	role       pbrole.RoleCommandServiceClient
	logger     logger.LoggerInterface
	mapper     apimapper.RoleCommandResponseMapper
	cache      role_cache.RoleCommandCache
	apiHandler apierror.ApiHandler
}

type roleCommandHandleDeps struct {
	client      pbrole.RoleCommandServiceClient
	queryClient pbrole.RoleQueryServiceClient
	userRole    pbuserrole.UserRoleServiceClient
	router      chi.Router
	logger      logger.LoggerInterface
	mapper      apimapper.RoleCommandResponseMapper
	kafka       *kafka.Kafka
	cache_role  mencache.RoleCache
	cache       role_cache.RoleCommandCache
	apiHandler  apierror.ApiHandler
}

func NewRoleCommandHandleApi(params *roleCommandHandleDeps) *roleCommandHandlerApi {
	handler := &roleCommandHandlerApi{
		role:       params.client,
		logger:     params.logger,
		mapper:     params.mapper,
		cache:      params.cache,
		apiHandler: params.apiHandler,
		kafka:      params.kafka,
	}

	roleMiddleware := middlewares.RoleValidatorGRPC(params.userRole, params.logger, params.cache_role)
	params.router.Route("/api/role-command", func(routerRole chi.Router) {
		requireAdmin := middlewares.RequireRoles("Admin", "ROLE_ADMIN", "Admin_Admin_14")

		routerRole.With(roleMiddleware, requireAdmin).Post("/create", httpx.Handler(handler.Create))
		routerRole.With(roleMiddleware, requireAdmin).Post("/update/{id}", httpx.Handler(handler.Update))
		routerRole.With(roleMiddleware, requireAdmin).Post("/trashed/{id}", httpx.Handler(handler.Trash))
		routerRole.With(roleMiddleware, requireAdmin).Post("/restore/{id}", httpx.Handler(handler.Restore))
		routerRole.With(roleMiddleware, requireAdmin).Delete("/permanent/{id}", httpx.Handler(handler.DeletePermanent))
		routerRole.With(roleMiddleware, requireAdmin).Post("/restore/all", httpx.Handler(handler.RestoreAll))
		routerRole.With(roleMiddleware, requireAdmin).Post("/permanent/all", httpx.Handler(handler.DeleteAllPermanent))

	})
	return handler
}

// @Security Bearer
// @Summary Create a new role
// @Tags Role Command
// @Description Create a new role with the provided details
// @Accept json
// @Produce json
// @Param body body requests.CreateRoleRequest true "Create role request"
// @Success 200 {object} response.ApiResponseRole "Successfully created role"
// @Failure 401 {object} errors.ErrorResponse "Unauthorized"
// @Failure 400 {object} errors.ErrorResponse "Invalid request parameters"
// @Failure 500 {object} errors.ErrorResponse "Failed to create role"
// @Router /api/role-command/create [post]
func (h *roleCommandHandlerApi) Create(w http.ResponseWriter, r *http.Request) error {
	var body requests.CreateRoleRequest
	if err := httpx.Bind(r, &body); err != nil {
		return errors.NewBadRequestError("Invalid request").WithInternal(err)
	}

	ctx := r.Context()
	res, err := h.role.CreateRole(ctx, &pbrole.CreateRoleRequest{Name: body.Name})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	return httpx.JSON(w, http.StatusOK, h.mapper.ToApiResponseRole(res))
}

// @Security Bearer
// @Summary Update role details
// @Tags Role Command
// @Description Update the name of an existing role
// @Accept json
// @Produce json
// @Param id path int true "Role ID"
// @Param body body requests.UpdateRoleRequest true "Update role request"
// @Success 200 {object} response.ApiResponseRole "Successfully updated role"
// @Failure 401 {object} errors.ErrorResponse "Unauthorized"
// @Failure 400 {object} errors.ErrorResponse "Invalid request parameters"
// @Failure 500 {object} errors.ErrorResponse "Failed to update role"
// @Router /api/role-command/update/{id} [post]
func (h *roleCommandHandlerApi) Update(w http.ResponseWriter, r *http.Request) error {
	roleID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || roleID <= 0 {
		return errors.NewBadRequestError("id is required")
	}

	var body requests.UpdateRoleRequest
	if err := httpx.Bind(r, &body); err != nil {
		return errors.NewBadRequestError("Invalid request").WithInternal(err)
	}

	ctx := r.Context()
	res, err := h.role.UpdateRole(ctx, &pbrole.UpdateRoleRequest{Id: int32(roleID), Name: body.Name})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	h.cache.DeleteCachedRole(ctx, roleID)

	return httpx.JSON(w, http.StatusOK, h.mapper.ToApiResponseRole(res))
}

// @Security Bearer
// @Summary Move role to trash
// @Tags Role Command
// @Description Move a role record to trash by its ID
// @Accept json
// @Produce json
// @Param id path int true "Role ID"
// @Success 200 {object} response.ApiResponseRole "Successfully moved role to trash"
// @Failure 400 {object} errors.ErrorResponse "Invalid role ID"
// @Failure 500 {object} errors.ErrorResponse "Failed to move role to trash"
// @Router /api/role-command/trashed/{id} [post]
func (h *roleCommandHandlerApi) Trash(w http.ResponseWriter, r *http.Request) error {
	roleID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || roleID <= 0 {
		return errors.NewBadRequestError("id is required")
	}

	ctx := r.Context()
	res, err := h.role.TrashedRole(ctx, &pbrole.FindByIdRoleRequest{RoleId: int32(roleID)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	h.cache.DeleteCachedRole(ctx, roleID)

	return httpx.JSON(w, http.StatusOK, h.mapper.ToApiResponseRole(res))
}

// @Security Bearer
// @Summary Restore a trashed role
// @Tags Role Command
// @Description Restore a trashed role record by its ID
// @Accept json
// @Produce json
// @Param id path int true "Role ID"
// @Success 200 {object} response.ApiResponseRole "Successfully restored role"
// @Failure 400 {object} errors.ErrorResponse "Invalid role ID"
// @Failure 500 {object} errors.ErrorResponse "Failed to restore role"
// @Router /api/role-command/restore/{id} [put]
func (h *roleCommandHandlerApi) Restore(w http.ResponseWriter, r *http.Request) error {
	roleID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || roleID <= 0 {
		return errors.NewBadRequestError("id is required")
	}

	ctx := r.Context()
	res, err := h.role.RestoreRole(ctx, &pbrole.FindByIdRoleRequest{RoleId: int32(roleID)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	h.cache.DeleteCachedRole(ctx, roleID)

	return httpx.JSON(w, http.StatusOK, h.mapper.ToApiResponseRole(res))
}

// @Security Bearer
// @Summary Permanently delete a role
// @Tags Role Command
// @Description Permanently delete a role record by its ID
// @Accept json
// @Produce json
// @Param id path int true "Role ID"
// @Success 200 {object} response.ApiResponseRoleDelete "Successfully deleted role record permanently"
// @Failure 400 {object} errors.ErrorResponse "Invalid role ID"
// @Failure 500 {object} errors.ErrorResponse "Failed to delete role permanently"
// @Router /api/role-command/permanent/{id} [delete]
func (h *roleCommandHandlerApi) DeletePermanent(w http.ResponseWriter, r *http.Request) error {
	roleID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || roleID <= 0 {
		return errors.NewBadRequestError("id is required")
	}

	ctx := r.Context()
	res, err := h.role.DeleteRolePermanent(ctx, &pbrole.FindByIdRoleRequest{RoleId: int32(roleID)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	h.cache.DeleteCachedRole(ctx, roleID)

	return httpx.JSON(w, http.StatusOK, h.mapper.ToApiResponseRoleDelete(res))
}

// @Security Bearer
// @Summary Restore all trashed roles
// @Tags Role Command
// @Description Restore all trashed role records
// @Accept json
// @Produce json
// @Success 200 {object} response.ApiResponseRoleAll "Successfully restored all roles"
// @Failure 500 {object} errors.ErrorResponse "Failed to restore roles"
// @Router /api/role-command/restore/all [post]
func (h *roleCommandHandlerApi) RestoreAll(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	res, err := h.role.RestoreAllRole(ctx, &emptypb.Empty{})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	return httpx.JSON(w, http.StatusOK, h.mapper.ToApiResponseRoleAll(res))
}

// @Security Bearer
// @Summary Permanently delete all trashed roles
// @Tags Role Command
// @Description Permanently delete all trashed role records
// @Accept json
// @Produce json
// @Success 200 {object} response.ApiResponseRoleAll "Successfully deleted all roles permanently"
// @Failure 500 {object} errors.ErrorResponse "Failed to delete roles permanently"
// @Router /api/role-command/permanent/all [delete]
func (h *roleCommandHandlerApi) DeleteAllPermanent(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	res, err := h.role.DeleteAllRolePermanent(ctx, &emptypb.Empty{})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	return httpx.JSON(w, http.StatusOK, h.mapper.ToApiResponseRoleAll(res))
}
