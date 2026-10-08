package order_cache

import "github.com/MamangRust/microservice-ecommerce-shared/cache"

type orderMencache struct {
	OrderQueryCache
	OrderCommandCache
}

type OrderMencache interface {
	OrderQueryCache
	OrderCommandCache
}

func OrderNewMencache(cacheStore *cache.CacheStore) OrderMencache {
	return &orderMencache{
		OrderQueryCache:   NewOrderQueryCache(cacheStore),
		OrderCommandCache: NewOrderCommandCache(cacheStore),
	}
}
