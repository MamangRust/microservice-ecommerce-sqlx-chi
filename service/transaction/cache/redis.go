package cache

import "github.com/MamangRust/microservice-ecommerce-shared/cache"

type Mencache struct {
	TransactionQueryCache           TransactionQueryCache
	TransactionCommandCache         TransactionCommandCache
	// F5: legacy OLTP transaction stats caches removed.
}

type TransactionMencache interface {
	TransactionQueryCache
	TransactionCommandCache
}

func NewMencache(cacheStore *cache.CacheStore) *Mencache {
	return &Mencache{
		TransactionQueryCache:           NewTransactionQueryCache(cacheStore),
		TransactionCommandCache:         NewTransactionCommandCache(cacheStore),

	}
}
