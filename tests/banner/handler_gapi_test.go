package banner_test

import (
	"context"
	"testing"

	banner_cache "github.com/MamangRust/microservice-ecommerce-grpc-banner/cache"
	banner_handler "github.com/MamangRust/microservice-ecommerce-grpc-banner/handler"
	banner_repo "github.com/MamangRust/microservice-ecommerce-grpc-banner/repository"
	banner_service "github.com/MamangRust/microservice-ecommerce-grpc-banner/service"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/banner"
	"github.com/MamangRust/microservice-ecommerce-shared/cache"
	"github.com/MamangRust/microservice-ecommerce-shared/observability"
	tests "github.com/MamangRust/microservice-ecommerce-test"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type BannerGapiTestSuite struct {
	tests.BaseTestSuite
	queryClient   pb_banner.BannerQueryServiceClient
	commandClient pb_banner.BannerCommandServiceClient
}

func (s *BannerGapiTestSuite) SetupSuite() {
	s.BaseTestSuite.SetupSuite()

	// Infrastructure
	cacheMetrics, _ := observability.NewCacheMetrics("test")
	cacheStore := cache.NewCacheStore(s.RedisClient(), s.Log, cacheMetrics)
	// Banner dependencies
	mencache := banner_cache.NewMencache(cacheStore)
	repos := banner_repo.NewRepositories(s.SQLxDB())
	svc := banner_service.NewService(&banner_service.Deps{
		Cache:         mencache,
		Repository:    repos,
		Logger:        s.Log,
		Observability: s.Obs,
	})

	// Handler
	handler := banner_handler.NewHandler(&banner_handler.Deps{
		Service: svc,
		Logger:  s.Log,
	})

	// Server
	server := grpc.NewServer()
	pb_banner.RegisterBannerQueryServiceServer(server, handler.BannerQuery)
	pb_banner.RegisterBannerCommandServiceServer(server, handler.BannerCommand)

	addr := s.RegisterServer(server)
	conn := s.GetConnection(addr)

	s.queryClient = pb_banner.NewBannerQueryServiceClient(conn)
	s.commandClient = pb_banner.NewBannerCommandServiceClient(conn)
}

func (s *BannerGapiTestSuite) TestBannerGapiLifecycle() {
	ctx := context.Background()

	// 1. Create
	createRes, err := s.commandClient.Create(ctx, &pb_banner.CreateBannerRequest{
		Name:      "GAPI Sale",
		StartDate: "2026-01-01",
		EndDate:   "2026-12-31",
		StartTime: "00:00:00",
		EndTime:   "23:59:59",
		IsActive:  true,
	})
	s.Require().NoError(err)
	s.Require().NotNil(createRes)
	bannerID := createRes.Data.BannerId

	// 2. FindById
	getRes, err := s.queryClient.FindById(ctx, &pb_banner.FindByIdBannerRequest{Id: bannerID})
	s.Require().NoError(err)
	s.Equal("GAPI Sale", getRes.Data.Name)

	// 3. FindAll
	allRes, err := s.queryClient.FindAll(ctx, &pb_banner.FindAllBannerRequest{Page: 1, PageSize: 10})
	s.Require().NoError(err)
	s.NotEmpty(allRes.Data)

	// 4. FindByActive
	activeRes, err := s.queryClient.FindByActive(ctx, &pb_banner.FindAllBannerRequest{Page: 1, PageSize: 10})
	s.Require().NoError(err)
	s.NotEmpty(activeRes.Data)

	// 5. Update
	updateRes, err := s.commandClient.Update(ctx, &pb_banner.UpdateBannerRequest{
		BannerId:  bannerID,
		Name:      "GAPI Sale Updated",
		StartDate: "2026-01-01",
		EndDate:   "2026-12-31",
		StartTime: "00:00:00",
		EndTime:   "23:59:59",
		IsActive:  true,
	})
	s.Require().NoError(err)
	s.Equal("GAPI Sale Updated", updateRes.Data.Name)

	// 6. Trash
	_, err = s.commandClient.Trash(ctx, &pb_banner.FindByIdBannerRequest{Id: bannerID})
	s.Require().NoError(err)

	// 7. FindByTrashed
	trashedRes, err := s.queryClient.FindByTrashed(ctx, &pb_banner.FindAllBannerRequest{Page: 1, PageSize: 10})
	s.Require().NoError(err)
	s.NotEmpty(trashedRes.Data)

	// 8. Restore
	_, err = s.commandClient.Restore(ctx, &pb_banner.FindByIdBannerRequest{Id: bannerID})
	s.Require().NoError(err)

	// 9. DeletePermanent
	_, _ = s.commandClient.Trash(ctx, &pb_banner.FindByIdBannerRequest{Id: bannerID})
	_, err = s.commandClient.DeletePermanent(ctx, &pb_banner.FindByIdBannerRequest{Id: bannerID})
	s.Require().NoError(err)

	// 10. RestoreAll
	_, err = s.commandClient.RestoreAll(ctx, &emptypb.Empty{})
	s.Require().NoError(err)

	// 11. DeleteAll
	_, err = s.commandClient.DeleteAll(ctx, &emptypb.Empty{})
	s.Require().NoError(err)
}

func TestBannerGapiSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(BannerGapiTestSuite))
}
