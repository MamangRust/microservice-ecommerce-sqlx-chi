package merchantdocumenthandler

import (
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	"github.com/MamangRust/microservice-ecommerce-pkg/upload_image"
	apimapper "github.com/MamangRust/microservice-ecommerce-shared/mapper/merchant_documents"
	pbmerchantdocument "github.com/MamangRust/microservice-ecommerce-grpc-pb/merchant_document"
	"github.com/go-chi/chi/v5"
	"google.golang.org/grpc"
)

type DepsMerchantDocument struct {
	Client      *grpc.ClientConn
	Router      chi.Router
	Logger      logger.LoggerInterface
	UploadImage upload_image.ImageUploads
}

func RegisterMerchantDocumentHandler(deps *DepsMerchantDocument) {
	mapper := apimapper.NewMerchantDocumentResponseMapper()

	NewMerchantDocumentQueryHandleApi(&merchantDocumentQueryHandleDeps{
		client: pbmerchantdocument.NewMerchantDocumentQueryServiceClient(deps.Client),
		router: deps.Router,
		logger: deps.Logger,
		mapper: mapper.QueryMapper(),
	})

	NewMerchantDocumentCommandHandleApi(&merchantDocumentCommandHandleDeps{
		client:       pbmerchantdocument.NewMerchantDocumentCommandServiceClient(deps.Client),
		router:       deps.Router,
		logger:       deps.Logger,
		mapper:       mapper.CommandMapper(),
		upload_image: deps.UploadImage,
	})
}
