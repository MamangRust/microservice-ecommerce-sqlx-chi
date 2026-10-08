package user_test

import (
	"context"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/role"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/user"
	pbuserrole "github.com/MamangRust/microservice-ecommerce-grpc-pb/user_role"
	gapi "github.com/MamangRust/microservice-ecommerce-grpc-user/handler"
	"github.com/MamangRust/microservice-ecommerce-grpc-user/repository"
	"github.com/MamangRust/microservice-ecommerce-grpc-user/service"
	"github.com/MamangRust/microservice-ecommerce-pkg/adapter"
	"github.com/MamangRust/microservice-ecommerce-pkg/hash"
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	"github.com/MamangRust/microservice-ecommerce-pkg/resilience"
	tests "github.com/MamangRust/microservice-ecommerce-test"
	"testing"
	"time"

	user_cache "github.com/MamangRust/microservice-ecommerce-grpc-user/cache"
	"github.com/stretchr/testify/suite"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type UserGapiTestSuite struct {
	tests.BaseTestSuite
	client      pb_user.UserCommandServiceClient
	queryClient pb_user.UserQueryServiceClient
	userID      int
}

func (s *UserGapiTestSuite) SetupSuite() {
	s.BaseTestSuite.SetupSuite()

	s.SetupRoleService()
	roleQueryClient := pb_role.NewRoleQueryServiceClient(s.Conns["role"])
	roleCommandClient := pbuserrole.NewUserRoleServiceClient(s.Conns["role"])

	guardRole := resilience.NewDependencyGuard("role", 5, 30, 100, 3*time.Second, s.Log)
	guardUserRole := resilience.NewDependencyGuard("user_role", 5, 30, 100, 3*time.Second, s.Log)

	repos := repository.NewRepositories(&repository.Deps{
		Db:       s.SQLxDB(),
		Role:     roleQueryClient,
		UserRole: roleCommandClient,
		Guards: repository.GuardOptions{
			Role:     []adapter.GuardOption{adapter.WithDependencyGuard(guardRole)},
			UserRole: []adapter.GuardOption{adapter.WithDependencyGuard(guardUserRole)},
		},
	})

	logger.ResetInstance()
	lp := sdklog.NewLoggerProvider()
	log, _ := logger.NewLogger("test", lp)
	hasher := hash.NewHashingPassword()
	cacheStore := s.GetCacheStore()
	mencache := user_cache.NewMencache(cacheStore)

	userService := service.NewService(&service.Deps{
		Repositories:  repos,
		Logger:        log,
		Hash:          hasher,
		Cache:         mencache,
		Observability: s.Obs,
	})

	// Start gRPC Server
	userHandler := gapi.NewHandler(&gapi.Deps{
		Service: userService,
		Logger:  log,
	})
	server := grpc.NewServer()
	pb_user.RegisterUserCommandServiceServer(server, userHandler.UserCommand)
	pb_user.RegisterUserQueryServiceServer(server, userHandler.UserQuery)

	addr := s.RegisterServer(server)
	conn := s.GetConnection(addr)

	s.client = pb_user.NewUserCommandServiceClient(conn)
	s.queryClient = pb_user.NewUserQueryServiceClient(conn)
}

func (s *UserGapiTestSuite) TestUserGapiLifecycle() {
	ctx := context.Background()

	// 1. Create
	createReq := &pb_user.CreateUserRequest{
		Firstname:       "Gapi",
		Lastname:        "User",
		Email:           "gapi.user@example.com",
		Password:        "password123",
		ConfirmPassword: "password123",
	}
	res, err := s.client.Create(ctx, createReq)
	s.Require().NoError(err)
	s.Equal(createReq.Email, res.Data.Email)
	userID := res.Data.Id

	// 2. FindById
	getRes, err := s.queryClient.FindById(ctx, &pb_user.FindByIdUserRequest{Id: userID})
	s.Require().NoError(err)
	s.Equal(userID, getRes.Data.Id)

	// 3. FindAll
	allRes, err := s.queryClient.FindAll(ctx, &pb_user.FindAllUserRequest{Page: 1, PageSize: 10})
	s.Require().NoError(err)
	s.NotEmpty(allRes.Data)

	// 4. FindByActive
	activeRes, err := s.queryClient.FindByActive(ctx, &pb_user.FindAllUserRequest{Page: 1, PageSize: 10})
	s.Require().NoError(err)
	s.NotEmpty(activeRes.Data)

	// 5. Update
	updateRes, err := s.client.Update(ctx, &pb_user.UpdateUserRequest{
		Id:              userID,
		Firstname:       "GapiUpdated",
		Lastname:        "UserUpdated",
		Email:           "gapi.updated@example.com",
		Password:        "password123",
		ConfirmPassword: "password123",
	})
	s.Require().NoError(err)
	s.Equal("GapiUpdated", updateRes.Data.Firstname)

	// 6. Trash
	_, err = s.client.TrashedUser(ctx, &pb_user.FindByIdUserRequest{Id: userID})
	s.Require().NoError(err)

	// 7. FindByTrashed
	trashedRes, err := s.queryClient.FindByTrashed(ctx, &pb_user.FindAllUserRequest{Page: 1, PageSize: 10})
	s.Require().NoError(err)
	s.NotEmpty(trashedRes.Data)

	// 8. Restore
	_, err = s.client.RestoreUser(ctx, &pb_user.FindByIdUserRequest{Id: userID})
	s.Require().NoError(err)

	// 9. DeletePermanent
	_, _ = s.client.TrashedUser(ctx, &pb_user.FindByIdUserRequest{Id: userID})
	_, err = s.client.DeleteUserPermanent(ctx, &pb_user.FindByIdUserRequest{Id: userID})
	s.Require().NoError(err)

	// 10. RestoreAll
	_, err = s.client.RestoreAllUser(ctx, &emptypb.Empty{})
	s.Require().NoError(err)

	// 11. DeleteAll
	_, err = s.client.DeleteAllUserPermanent(ctx, &emptypb.Empty{})
	s.Require().NoError(err)
}

func TestUserGapiSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(UserGapiTestSuite))
}
