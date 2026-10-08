package handler

import (
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/transaction"
)

type TransactionQueryHandler interface {
	pb_transaction.TransactionQueryServiceServer
}

type TransactionCommandHandler interface {
	pb_transaction.TransactionCommandServiceServer
}
