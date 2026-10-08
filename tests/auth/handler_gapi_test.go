package auth_test

import (
	"context"
	"net"
	"strconv"
	"testing"
	"time"

	mencache "github.com/MamangRust/microservice-ecommerce-auth/cache"
	"github.com/MamangRust/microservice-ecommerce-auth/handler"
	"github.com/MamangRust/microservice-ecommerce-auth/repository"
	"github.com/MamangRust/microservice-ecommerce-auth/service"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/auth"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/role"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/user"
	pbuserrole "github.com/MamangRust/microservice-ecommerce-grpc-pb/user_role"
	"github.com/MamangRust/microservice-ecommerce-pkg/adapter"
	"github.com/MamangRust/microservice-ecommerce-pkg/auth"
	"github.com/MamangRust/microservice-ecommerce-pkg/hash"
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	"github.com/MamangRust/microservice-ecommerce-pkg/resilience"
	"github.com/MamangRust/microservice-ecommerce-shared/cache"
	"github.com/MamangRust/microservice-ecommerce-shared/observability"
	tests "github.com/MamangRust/microservice-ecommerce-test"

	role_cache "github.com/MamangRust/microservice-ecommerce-grpc-role/cache"
	role_handler "github.com/MamangRust/microservice-ecommerce-grpc-role/handler"
	role_repo "github.com/MamangRust/microservice-ecommerce-grpc-role/repository"
	role_service "github.com/MamangRust/microservice-ecommerce-grpc-role/service"
	user_cache "github.com/MamangRust/microservice-ecommerce-grpc-user/cache"
	user_handler "github.com/MamangRust/microservice-ecommerce-grpc-user/handler"
	user_repo "github.com/MamangRust/microservice-ecommerce-grpc-user/repository"
	user_service "github.com/MamangRust/microservice-ecommerce-grpc-user/service"

	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type AuthHandlerGapiTestSuite struct {
	suite.Suite
	ts          *tests.TestSuite
	dbPool      *sqlx.DB
	redisClient *redis.Client
	client      pb_auth.AuthServiceClient
	conn        *grpc.ClientConn
	grpcServer  *grpc.Server
	email       string
	password    string
	accessToken string
}

func (s *AuthHandlerGapiTestSuite) SetupSuite() {
	ts, err := tests.SetupTestSuite()
	s.Require().NoError(err)
	s.ts = ts

	pool, err := sqlx.ConnectContext(s.ts.Ctx, "pgx", s.ts.DBURL)
	s.Require().NoError(err)
	s.dbPool = pool

	opts, err := redis.ParseURL(s.ts.RedisURL)
	s.Require().NoError(err)
	s.redisClient = redis.NewClient(opts)

	queries := pool

	logger.ResetInstance()
	lp := sdklog.NewLoggerProvider()
	log, _ := logger.NewLogger("test", lp)
	hasher := hash.NewHashingPassword()
	cacheMetrics, _ := observability.NewCacheMetrics("test")
	cacheStore := cache.NewCacheStore(s.redisClient, log, cacheMetrics)
	obs, _ := observability.NewObservability("test", log)

	// 1. Setup Role Service & gRPC Server
	roleMencache := role_cache.NewMencache(cacheStore)
	roleRepos := role_repo.NewRepositories(pool)
	roleSvc := role_service.NewService(&role_service.Deps{
		Repository:    roleRepos,
		Logger:        log,
		Cache:         roleMencache,
		Observability: obs,
	})
	roleGapi := role_handler.NewHandler(&role_handler.Deps{
		Service: roleSvc,
		Logger:  log,
	})
	roleServer := grpc.NewServer()
	pb_role.RegisterRoleQueryServiceServer(roleServer, roleGapi.RoleQuery)
	pb_role.RegisterRoleCommandServiceServer(roleServer, roleGapi.RoleCommand)
	pbuserrole.RegisterUserRoleServiceServer(roleServer, roleGapi.UserRole)
	roleLis, _ := net.Listen("tcp", "localhost:0")
	go roleServer.Serve(roleLis)
	roleConn, _ := grpc.NewClient(roleLis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))

	// 2. Setup User Service & gRPC Server
	userMencache := user_cache.NewMencache(cacheStore)
	roleQueryClientForUser := pb_role.NewRoleQueryServiceClient(roleConn)
	userRoleClientForUser := pbuserrole.NewUserRoleServiceClient(roleConn)
	guardUserRepoRole := resilience.NewDependencyGuard("role", 5, 30, 100, 3*time.Second, log)
	guardUserRepoUserRole := resilience.NewDependencyGuard("user_role", 5, 30, 100, 3*time.Second, log)
	userRepos := user_repo.NewRepositories(&user_repo.Deps{
		Db:       pool,
		Role:     roleQueryClientForUser,
		UserRole: userRoleClientForUser,
		Guards: user_repo.GuardOptions{
			Role:     []adapter.GuardOption{adapter.WithDependencyGuard(guardUserRepoRole)},
			UserRole: []adapter.GuardOption{adapter.WithDependencyGuard(guardUserRepoUserRole)},
		},
	})
	userSvc := user_service.NewService(&user_service.Deps{
		Repositories:  userRepos,
		Logger:        log,
		Hash:          hasher,
		Cache:         userMencache,
		Observability: obs,
	})
	userGapi := user_handler.NewHandler(&user_handler.Deps{
		Service: userSvc,
		Logger:  log,
	})
	userServer := grpc.NewServer()
	pb_user.RegisterUserQueryServiceServer(userServer, userGapi.UserQuery)
	pb_user.RegisterUserCommandServiceServer(userServer, userGapi.UserCommand)
	userLis, _ := net.Listen("tcp", "localhost:0")
	go userServer.Serve(userLis)
	userConn, _ := grpc.NewClient(userLis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))

	// 3. Setup Auth Service with gRPC clients
	userQueryClient := pb_user.NewUserQueryServiceClient(userConn)
	userCommandClient := pb_user.NewUserCommandServiceClient(userConn)
	roleQueryClient := pb_role.NewRoleQueryServiceClient(roleConn)
	roleCommandClient := pbuserrole.NewUserRoleServiceClient(roleConn)

	guardAuthUser := resilience.NewDependencyGuard("user", 5, 30, 100, 3*time.Second, log)
	guardAuthRole := resilience.NewDependencyGuard("role", 5, 30, 100, 3*time.Second, log)
	guardAuthUserRole := resilience.NewDependencyGuard("user_role", 5, 30, 100, 3*time.Second, log)

	repos := repository.NewRepositories(&repository.Deps{
		Db:              queries,
		User:            userQueryClient,
		UserCommand:     userCommandClient,
		Role:            roleQueryClient,
		UserRoleCommand: roleCommandClient,
		Guards: repository.GuardOptions{
			User:     []adapter.GuardOption{adapter.WithDependencyGuard(guardAuthUser)},
			Role:     []adapter.GuardOption{adapter.WithDependencyGuard(guardAuthRole)},
			UserRole: []adapter.GuardOption{adapter.WithDependencyGuard(guardAuthUserRole)},
		},
	})

	tokenManager, _ := auth.NewManager("mysecret")
	svc := service.NewService(&service.Deps{
		Repositories:  repos,
		Logger:        log,
		Mencache:      mencache.NewMencache(cacheStore),
		Token:         tokenManager,
		Hash:          hasher,
		Kafka:         nil,
		Observability: obs,
	})

	h := handler.NewAuthHandleGrpc(svc, log)

	s.grpcServer = grpc.NewServer()
	pb_auth.RegisterAuthServiceServer(s.grpcServer, h)

	lis, err := net.Listen("tcp", "localhost:0")
	s.Require().NoError(err)

	go func() {
		_ = s.grpcServer.Serve(lis)
	}()

	conn, err := grpc.NewClient(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	s.Require().NoError(err)
	s.conn = conn
	s.client = pb_auth.NewAuthServiceClient(conn)

	s.email = "auth.handler.gapi.test@example.com"
	s.password = "password123"
}

func (s *AuthHandlerGapiTestSuite) TearDownSuite() {
	if s.conn != nil {
		s.conn.Close()
	}
	if s.grpcServer != nil {
		s.grpcServer.Stop()
	}
	if s.redisClient != nil {
		s.redisClient.Close()
	}
	if s.dbPool != nil {
		s.dbPool.Close()
	}
	s.ts.Teardown()
}

func (s *AuthHandlerGapiTestSuite) Test1_Register() {
	ctx := context.Background()
	req := &pb_auth.RegisterRequest{
		Firstname:       "Auth",
		Lastname:        "Handler",
		Email:           s.email,
		Password:        s.password,
		ConfirmPassword: s.password,
	}

	res, err := s.client.RegisterUser(ctx, req)
	s.NoError(err)
	s.NotNil(res)
	s.Equal("success", res.Status)
	s.Equal(s.email, res.Data.Email)
}

func (s *AuthHandlerGapiTestSuite) Test2_Login() {
	ctx := context.Background()
	req := &pb_auth.LoginRequest{
		Email:    s.email,
		Password: s.password,
	}

	res, err := s.client.LoginUser(ctx, req)
	s.NoError(err)
	s.NotNil(res)
	s.Equal("success", res.Status)
	s.NotEmpty(res.Data.AccessToken)
	s.accessToken = res.Data.AccessToken
}

func (s *AuthHandlerGapiTestSuite) Test4_LoginLockout() {
	ctx := context.Background()
	email := "locked.gapi@example.com"
	password := "wrongpassword"

	// Register user first
	regReq := &pb_auth.RegisterRequest{
		Firstname:       "Locked",
		Lastname:        "Gapi",
		Email:           email,
		Password:        "correctpassword",
		ConfirmPassword: "correctpassword",
	}
	_, err := s.client.RegisterUser(ctx, regReq)
	s.NoError(err)

	loginReq := &pb_auth.LoginRequest{
		Email:    email,
		Password: password,
	}

	// Fail login 5 times
	for i := 0; i < 5; i++ {
		_, err := s.client.LoginUser(ctx, loginReq)
		s.Error(err)
	}

	// 6th attempt should return error
	_, err = s.client.LoginUser(ctx, loginReq)
	s.Error(err)
	s.Contains(err.Error(), "Account temporarily locked")
}

func (s *AuthHandlerGapiTestSuite) Test3_GetMe() {
	s.Require().NotEmpty(s.accessToken)
	ctx := context.Background()

	tokenManager, _ := auth.NewManager("mysecret")
	userIdStr, err := tokenManager.ValidateToken(s.accessToken)
	s.NoError(err)

	userId, err := strconv.Atoi(userIdStr)
	s.NoError(err)

	res, err := s.client.GetMe(ctx, &pb_auth.GetMeRequest{UserId: int32(userId)})
	s.NoError(err)
	s.NotNil(res)
	s.Equal("success", res.Status)
	s.Equal(s.email, res.Data.Email)
}

func TestAuthHandlerGapiSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(AuthHandlerGapiTestSuite))
}
