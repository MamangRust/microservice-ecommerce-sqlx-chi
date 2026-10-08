package apps

import (
	"fmt"
	"time"

	pb_role "github.com/MamangRust/microservice-ecommerce-grpc-pb/role"
	pb_user "github.com/MamangRust/microservice-ecommerce-grpc-pb/user"
	pbuserrole "github.com/MamangRust/microservice-ecommerce-grpc-pb/user_role"
	"github.com/MamangRust/microservice-ecommerce-grpc-user/cache"
	"github.com/MamangRust/microservice-ecommerce-grpc-user/handler"
	"github.com/MamangRust/microservice-ecommerce-grpc-user/repository"
	"github.com/MamangRust/microservice-ecommerce-grpc-user/service"
	"github.com/MamangRust/microservice-ecommerce-pkg/adapter"
	"github.com/MamangRust/microservice-ecommerce-pkg/hash"
	"github.com/MamangRust/microservice-ecommerce-pkg/resilience"
	"github.com/MamangRust/microservice-ecommerce-pkg/server"
	"github.com/MamangRust/microservice-ecommerce-shared/observability"
	"github.com/spf13/viper"
	"google.golang.org/grpc"

	pkgresilience "github.com/MamangRust/microservice-ecommerce-pkg/resilience"
	"google.golang.org/grpc/credentials/insecure"
)

func NewServer(cfg *server.Config) (*server.GRPCServer, error) {
	srv, err := server.New(cfg)
	if err != nil {
		return nil, err
	}

	roleAddr := viper.GetString("GRPC_ROLE_ADDR")

	roleConn, err := grpc.NewClient(
		roleAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(pkgresilience.NewDependencyGuardInterceptor(srv.Logger).UnaryInterceptor()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to role service: %w", err)
	}

	roleQueryClient := pb_role.NewRoleQueryServiceClient(roleConn)
	userRoleClient := pbuserrole.NewUserRoleServiceClient(roleConn)

	repos := repository.NewRepositories(&repository.Deps{
		Db:       srv.DB,
		Role:     roleQueryClient,
		UserRole: userRoleClient,
		Guards: repository.GuardOptions{
			Role: []adapter.GuardOption{
				adapter.WithDependencyGuard(resilience.NewDependencyGuard("role", 5, 30, 100, 3*time.Second, srv.Logger)),
			},
			UserRole: []adapter.GuardOption{
				adapter.WithDependencyGuard(resilience.NewDependencyGuard("user_role", 5, 30, 100, 3*time.Second, srv.Logger)),
			},
		},
	})

	hashing := hash.NewHashingPassword()
	obs, _ := observability.NewObservability("user-server", srv.Logger)
	cache := cache.NewMencache(srv.CacheStore)

	svc := service.NewService(&service.Deps{
		Cache:         cache,
		Repositories:  repos,
		Hash:          hashing,
		Logger:        srv.Logger,
		Observability: obs,
	})

	h := handler.NewHandler(&handler.Deps{
		Service: svc,
		Logger:  srv.Logger,
	})

	srv.RegisterServices = func(gs *grpc.Server) {
		pb_user.RegisterUserQueryServiceServer(gs, h.UserQuery)
		pb_user.RegisterUserCommandServiceServer(gs, h.UserCommand)
	}

	return srv, nil
}
