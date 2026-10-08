package role_test

import (
	"context"
	"testing"

	"github.com/MamangRust/microservice-ecommerce-grpc-pb/role"
	role_cache "github.com/MamangRust/microservice-ecommerce-grpc-role/cache"
	role_handler "github.com/MamangRust/microservice-ecommerce-grpc-role/handler"
	"github.com/MamangRust/microservice-ecommerce-grpc-role/repository"
	"github.com/MamangRust/microservice-ecommerce-grpc-role/service"
	"github.com/MamangRust/microservice-ecommerce-shared/cache"
	"github.com/MamangRust/microservice-ecommerce-shared/observability"
	tests "github.com/MamangRust/microservice-ecommerce-test"

	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type RoleGapiTestSuite struct {
	tests.BaseTestSuite
	commandClient pb_role.RoleCommandServiceClient
	queryClient   pb_role.RoleQueryServiceClient
}

func (s *RoleGapiTestSuite) SetupSuite() {
	s.BaseTestSuite.SetupSuite()

	// Infrastructure
	cacheMetrics, _ := observability.NewCacheMetrics("test")
	cacheStore := cache.NewCacheStore(s.RedisClient(), s.Log, cacheMetrics)
	// Role dependencies
	mencache := role_cache.NewMencache(cacheStore)
	repos := repository.NewRepositories(s.SQLxDB())
	svc := service.NewService(&service.Deps{
		Repository:    repos,
		Logger:        s.Log,
		Cache:         mencache,
		Observability: s.Obs,
	})

	// Handler
	handler := role_handler.NewHandler(&role_handler.Deps{
		Service: svc,
		Logger:  s.Log,
	})

	// Server
	server := grpc.NewServer()
	pb_role.RegisterRoleCommandServiceServer(server, handler.RoleCommand)
	pb_role.RegisterRoleQueryServiceServer(server, handler.RoleQuery)

	addr := s.RegisterServer(server)
	conn := s.GetConnection(addr)

	s.commandClient = pb_role.NewRoleCommandServiceClient(conn)
	s.queryClient = pb_role.NewRoleQueryServiceClient(conn)
}

func (s *RoleGapiTestSuite) TestRoleGapiLifecycle() {
	ctx := context.Background()

	// 1. Create
	createRes, err := s.commandClient.CreateRole(ctx, &pb_role.CreateRoleRequest{
		Name: "Gapi Role",
	})
	s.Require().NoError(err)
	s.Require().NotNil(createRes)
	roleID := createRes.Data.Id

	// 2. FindById
	getRes, err := s.queryClient.FindByIdRole(ctx, &pb_role.FindByIdRoleRequest{RoleId: roleID})
	s.Require().NoError(err)
	s.Equal("Gapi Role", getRes.Data.Name)

	// 3. FindAll
	allRes, err := s.queryClient.FindAllRole(ctx, &pb_role.FindAllRoleRequest{Page: 1, PageSize: 10})
	s.Require().NoError(err)
	s.NotEmpty(allRes.Data)

	// 4. FindByActive
	activeRes, err := s.queryClient.FindByActive(ctx, &pb_role.FindAllRoleRequest{Page: 1, PageSize: 10})
	s.Require().NoError(err)
	s.NotEmpty(activeRes.Data)

	// 5. Update
	updateRes, err := s.commandClient.UpdateRole(ctx, &pb_role.UpdateRoleRequest{
		Id:   roleID,
		Name: "Gapi Role Updated",
	})
	s.Require().NoError(err)
	s.Equal("Gapi Role Updated", updateRes.Data.Name)

	// 6. Trash
	_, err = s.commandClient.TrashedRole(ctx, &pb_role.FindByIdRoleRequest{RoleId: roleID})
	s.Require().NoError(err)

	// 7. FindByTrashed
	trashedRes, err := s.queryClient.FindByTrashed(ctx, &pb_role.FindAllRoleRequest{Page: 1, PageSize: 10})
	s.Require().NoError(err)
	s.NotEmpty(trashedRes.Data)

	// 8. Restore
	_, err = s.commandClient.RestoreRole(ctx, &pb_role.FindByIdRoleRequest{RoleId: roleID})
	s.Require().NoError(err)

	// 9. DeletePermanent
	_, _ = s.commandClient.TrashedRole(ctx, &pb_role.FindByIdRoleRequest{RoleId: roleID})
	_, err = s.commandClient.DeleteRolePermanent(ctx, &pb_role.FindByIdRoleRequest{RoleId: roleID})
	s.Require().NoError(err)

	// 10. RestoreAll
	_, err = s.commandClient.RestoreAllRole(ctx, &emptypb.Empty{})
	s.Require().NoError(err)

	// 11. DeleteAll
	_, err = s.commandClient.DeleteAllRolePermanent(ctx, &emptypb.Empty{})
	s.Require().NoError(err)
}

func TestRoleGapiSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(RoleGapiTestSuite))
}
