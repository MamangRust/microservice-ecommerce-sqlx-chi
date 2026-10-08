package cache

import (
	"github.com/MamangRust/microservice-ecommerce-shared/cache"
)

// F5: legacy OLTP order stats caches were removed; stats are served by
// service/stats_reader from ClickHouse.
type Mencache struct {
	OrderQueryCache   OrderQueryCache
	OrderCommandCache OrderCommandCache
}

func NewMencache(cacheStore *cache.CacheStore) *Mencache {
	return &Mencache{
		OrderQueryCache:   NewOrderQueryCache(cacheStore),
		OrderCommandCache: NewOrderCommandCache(cacheStore),
	}
}
