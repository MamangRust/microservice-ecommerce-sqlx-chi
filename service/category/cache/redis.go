package cache

import (
	"github.com/MamangRust/microservice-ecommerce-shared/cache"
)

// F5: legacy OLTP category stats caches were removed; stats are served by
// service/stats_reader from ClickHouse.
type Mencache struct {
	CategoryQueryCache   CategoryQueryCache
	CategoryCommandCache CategoryCommandCache
}

func NewMencache(cacheStore *cache.CacheStore) *Mencache {
	return &Mencache{
		CategoryQueryCache:   NewCategoryQueryCache(cacheStore),
		CategoryCommandCache: NewCategoryCommandCache(cacheStore),
	}
}
