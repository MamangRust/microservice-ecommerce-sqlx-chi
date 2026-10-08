package handler

import (
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/review"
)

type ReviewHandleGrpc interface {
	pb_review.ReviewQueryServiceServer
	pb_review.ReviewCommandServiceServer
}

type ReviewQueryHandler interface {
	pb_review.ReviewQueryServiceServer
}

type ReviewCommandHandler interface {
	pb_review.ReviewCommandServiceServer
}
