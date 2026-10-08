package category_cache

import "github.com/MamangRust/microservice-ecommerce-shared/cache"

type CategoryMencache interface {
	CategoryQueryCache
	CategoryCommandCache
}

type categoryMencache struct {
	CategoryQueryCache
	CategoryCommandCache
}

func NewCategoryMencache(cacheStore *cache.CacheStore) CategoryMencache {
	return &categoryMencache{
		CategoryQueryCache:   NewCategoryQueryCache(cacheStore),
		CategoryCommandCache: NewCategoryCommandCache(cacheStore),
	}
}
