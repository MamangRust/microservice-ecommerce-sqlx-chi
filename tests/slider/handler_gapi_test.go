package slider_test

import (
	"context"
	"testing"

	"github.com/MamangRust/microservice-ecommerce-grpc-pb/slider"
	slider_cache "github.com/MamangRust/microservice-ecommerce-grpc-slider/cache"
	slider_handler "github.com/MamangRust/microservice-ecommerce-grpc-slider/handler"
	slider_repo "github.com/MamangRust/microservice-ecommerce-grpc-slider/repository"
	slider_service "github.com/MamangRust/microservice-ecommerce-grpc-slider/service"
	"github.com/MamangRust/microservice-ecommerce-shared/cache"
	"github.com/MamangRust/microservice-ecommerce-shared/observability"
	tests "github.com/MamangRust/microservice-ecommerce-test"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type SliderGapiTestSuite struct {
	tests.BaseTestSuite
	queryClient   pb_slider.SliderQueryServiceClient
	commandClient pb_slider.SliderCommandServiceClient
}

func (s *SliderGapiTestSuite) SetupSuite() {
	s.BaseTestSuite.SetupSuite()

	// Infrastructure
	cacheMetrics, _ := observability.NewCacheMetrics("test")
	cacheStore := cache.NewCacheStore(s.RedisClient(), s.Log, cacheMetrics)
	// Slider dependencies
	mencache := slider_cache.NewMencache(cacheStore)
	repos := slider_repo.NewRepositories(s.SQLxDB())
	svc := slider_service.NewService(&slider_service.Deps{
		Mencache:      mencache,
		Repositories:  repos,
		Logger:        s.Log,
		Observability: s.Obs,
	})

	// Handler
	handler := slider_handler.NewHandler(&slider_handler.Deps{
		Service: svc,
		Logger:  s.Log,
	})

	// Server
	server := grpc.NewServer()
	pb_slider.RegisterSliderQueryServiceServer(server, handler.SliderQuery)
	pb_slider.RegisterSliderCommandServiceServer(server, handler.SliderCommand)

	addr := s.RegisterServer(server)
	conn := s.GetConnection(addr)

	s.queryClient = pb_slider.NewSliderQueryServiceClient(conn)
	s.commandClient = pb_slider.NewSliderCommandServiceClient(conn)
}

func (s *SliderGapiTestSuite) TestSliderGapiLifecycle() {
	ctx := context.Background()

	// 1. Create
	createRes, err := s.commandClient.Create(ctx, &pb_slider.CreateSliderRequest{
		Name:  "GAPI Slider",
		Image: "http://example.com/gapi-slider.jpg",
	})
	s.Require().NoError(err)
	s.Require().NotNil(createRes)
	sliderID := createRes.Data.Id

	// 2. FindById
	getRes, err := s.queryClient.FindById(ctx, &pb_slider.FindByIdSliderRequest{Id: sliderID})
	s.Require().NoError(err)
	s.Equal("GAPI Slider", getRes.Data.Name)

	// 3. FindAll
	allRes, err := s.queryClient.FindAll(ctx, &pb_slider.FindAllSliderRequest{Page: 1, PageSize: 10})
	s.Require().NoError(err)
	s.NotEmpty(allRes.Data)

	// 4. FindByActive
	activeRes, err := s.queryClient.FindByActive(ctx, &pb_slider.FindAllSliderRequest{Page: 1, PageSize: 10})
	s.Require().NoError(err)
	s.NotEmpty(activeRes.Data)

	// 5. Update
	updateRes, err := s.commandClient.Update(ctx, &pb_slider.UpdateSliderRequest{
		Id:    sliderID,
		Name:  "GAPI Slider Updated",
		Image: "http://example.com/gapi-slider-updated.jpg",
	})
	s.Require().NoError(err)
	s.Equal("GAPI Slider Updated", updateRes.Data.Name)

	// 6. Trash
	_, err = s.commandClient.TrashedSlider(ctx, &pb_slider.FindByIdSliderRequest{Id: sliderID})
	s.Require().NoError(err)

	// 7. FindByTrashed
	trashedRes, err := s.queryClient.FindByTrashed(ctx, &pb_slider.FindAllSliderRequest{Page: 1, PageSize: 10})
	s.Require().NoError(err)
	s.NotEmpty(trashedRes.Data)

	// 8. Restore
	_, err = s.commandClient.RestoreSlider(ctx, &pb_slider.FindByIdSliderRequest{Id: sliderID})
	s.Require().NoError(err)

	// 9. DeletePermanent
	_, _ = s.commandClient.TrashedSlider(ctx, &pb_slider.FindByIdSliderRequest{Id: sliderID})
	_, err = s.commandClient.DeleteSliderPermanent(ctx, &pb_slider.FindByIdSliderRequest{Id: sliderID})
	s.Require().NoError(err)

	// 10. RestoreAll
	_, err = s.commandClient.RestoreAllSlider(ctx, &emptypb.Empty{})
	s.Require().NoError(err)

	// 11. DeleteAll
	_, err = s.commandClient.DeleteAllSliderPermanent(ctx, &emptypb.Empty{})
	s.Require().NoError(err)
}

func TestSliderGapiSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(SliderGapiTestSuite))
}
