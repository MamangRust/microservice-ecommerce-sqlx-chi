package transaction_cache

import "github.com/MamangRust/microservice-ecommerce-shared/cache"

type transactionMencache struct {
	TransactionQueryCache
	TransactionCommandCache
}

type TransactionMencache interface {
	TransactionQueryCache
	TransactionCommandCache
}

func NewTransactionMencache(cacheStore *cache.CacheStore) *transactionMencache {
	return &transactionMencache{
		TransactionQueryCache:   NewTransactionQueryCache(cacheStore),
		TransactionCommandCache: NewTransactionCommandCache(cacheStore),
	}
}
