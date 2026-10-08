package merchant_business_test

import (
	"context"
	"testing"
	"time"

	biz_cache "github.com/MamangRust/microservice-ecommerce-grpc-merchant_business/cache"
	biz_handler "github.com/MamangRust/microservice-ecommerce-grpc-merchant_business/handler"
	biz_repo "github.com/MamangRust/microservice-ecommerce-grpc-merchant_business/repository"
	biz_service "github.com/MamangRust/microservice-ecommerce-grpc-merchant_business/service"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/merchant"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/merchant_business"
	"github.com/MamangRust/microservice-ecommerce-pkg/adapter"
	"github.com/MamangRust/microservice-ecommerce-pkg/resilience"
	"github.com/MamangRust/microservice-ecommerce-shared/cache"
	"github.com/MamangRust/microservice-ecommerce-shared/observability"
	"github.com/MamangRust/microservice-ecommerce-test"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type MerchantBusinessGapiTestSuite struct {
	tests.BaseTestSuite
	queryClient   pb_merchant_business.MerchantBusinessQueryServiceClient
	commandClient pb_merchant_business.MerchantBusinessCommandServiceClient
}

func (s *MerchantBusinessGapiTestSuite) SetupSuite() {
	s.BaseTestSuite.SetupSuite()

	// Setup dependencies
	s.SetupUserService()
	s.SetupMerchantService()

	// Infrastructure
	cacheMetrics, _ := observability.NewCacheMetrics("test")
	cacheStore := cache.NewCacheStore(s.RedisClient(), s.Log, cacheMetrics)
	// Business dependencies
	mencache := biz_cache.NewMencache(cacheStore)
	guardMerchant := resilience.NewDependencyGuard("merchant", 5, 30, 100, 3*time.Second, s.Log)
	repos := biz_repo.NewRepositories(
		s.SQLxDB(),
		pb_merchant.NewMerchantQueryServiceClient(s.Conns["merchant"]),
		biz_repo.GuardOptions{
			Merchant: []adapter.GuardOption{adapter.WithDependencyGuard(guardMerchant)},
		},
	)
	svc := biz_service.NewService(&biz_service.Deps{
		Cache:         mencache,
		Repository:    repos,
		Logger:        s.Log,
		Observability: s.Obs,
	})

	// Handler
	handler := biz_handler.NewHandler(&biz_handler.Deps{
		Service: svc,
		Logger:  s.Log,
	})

	// Server
	server := grpc.NewServer()
	pb_merchant_business.RegisterMerchantBusinessQueryServiceServer(server, handler.MerchantBusinessQuery)
	pb_merchant_business.RegisterMerchantBusinessCommandServiceServer(server, handler.MerchantBusinessCommand)

	addr := s.RegisterServer(server)
	conn := s.GetConnection(addr)

	s.queryClient = pb_merchant_business.NewMerchantBusinessQueryServiceClient(conn)
	s.commandClient = pb_merchant_business.NewMerchantBusinessCommandServiceClient(conn)
}

func (s *MerchantBusinessGapiTestSuite) TestMerchantBusinessGapiLifecycle() {
	ctx := context.Background()

	// 1. Seed dependencies
	userID := s.SeedUser(ctx)
	merchID := int32(s.SeedMerchant(ctx, userID))

	// 2. Create
	createRes, err := s.commandClient.Create(ctx, &pb_merchant_business.CreateMerchantBusinessRequest{
		MerchantId:        merchID,
		BusinessType:      "GAPI Corp",
		TaxId:             "GAPI-TAX",
		EstablishedYear:   2020,
		NumberOfEmployees: 50,
		WebsiteUrl:        "http://gapi.corp",
	})
	s.NoError(err)
	s.NotNil(createRes)
	bizID := createRes.Data.Id

	// 3. FindById
	getRes, err := s.queryClient.FindById(ctx, &pb_merchant_business.FindByIdMerchantBusinessRequest{Id: bizID})
	s.NoError(err)
	s.Equal("GAPI Corp", getRes.Data.BusinessType)

	// 4. FindAll
	allRes, err := s.queryClient.FindAll(ctx, &pb_merchant.FindAllMerchantRequest{Page: 1, PageSize: 10})
	s.NoError(err)
	s.NotEmpty(allRes.Data)

	// 5. FindByActive
	activeRes, err := s.queryClient.FindByActive(ctx, &pb_merchant.FindAllMerchantRequest{Page: 1, PageSize: 10})
	s.NoError(err)
	s.NotEmpty(activeRes.Data)

	// 6. Update
	updateRes, err := s.commandClient.Update(ctx, &pb_merchant_business.UpdateMerchantBusinessRequest{
		MerchantBusinessInfoId: bizID,
		BusinessType:           "GAPI Corp Updated",
		TaxId:                  "GAPI-TAX-UPDATED",
		EstablishedYear:        2021,
		NumberOfEmployees:      60,
		WebsiteUrl:             "http://gapi.corp/updated",
	})
	s.NoError(err)
	s.Equal("GAPI Corp Updated", updateRes.Data.BusinessType)

	// 7. Trash
	_, err = s.commandClient.TrashedMerchantBusiness(ctx, &pb_merchant_business.FindByIdMerchantBusinessRequest{Id: bizID})
	s.NoError(err)

	// 8. FindByTrashed
	trashedRes, err := s.queryClient.FindByTrashed(ctx, &pb_merchant.FindAllMerchantRequest{Page: 1, PageSize: 10})
	s.NoError(err)
	s.NotEmpty(trashedRes.Data)

	// 9. Restore
	_, err = s.commandClient.RestoreMerchantBusiness(ctx, &pb_merchant_business.FindByIdMerchantBusinessRequest{Id: bizID})
	s.NoError(err)

	// 10. DeletePermanent
	_, _ = s.commandClient.TrashedMerchantBusiness(ctx, &pb_merchant_business.FindByIdMerchantBusinessRequest{Id: bizID})
	_, err = s.commandClient.DeleteMerchantBusinessPermanent(ctx, &pb_merchant_business.FindByIdMerchantBusinessRequest{Id: bizID})
	s.NoError(err)

	// 11. RestoreAll
	_, err = s.commandClient.RestoreAllMerchantBusiness(ctx, &emptypb.Empty{})
	s.NoError(err)

	// 12. DeleteAll
	_, err = s.commandClient.DeleteAllMerchantBusinessPermanent(ctx, &emptypb.Empty{})
	s.NoError(err)
}

func TestMerchantBusinessGapiSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(MerchantBusinessGapiTestSuite))
}
