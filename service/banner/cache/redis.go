package cache

import (
	sharedcachehelpers "github.com/MamangRust/microservice-ecommerce-shared/cache"
)

type Mencache struct {
	BannerQueryCache   BannerQueryCache
	BannerCommandCache BannerCommandCache
}

func NewMencache(cacheStore *sharedcachehelpers.CacheStore) *Mencache {
	return &Mencache{
		BannerQueryCache:   NewBannerQueryCache(cacheStore),
		BannerCommandCache: NewBannerCommandCache(cacheStore),
	}
}
