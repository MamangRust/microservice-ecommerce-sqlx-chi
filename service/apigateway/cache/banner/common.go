package banner_cache

import "time"

const (
	bannerAllCacheKey     = "apigw:banner:all:page:%d:pageSize:%d:search:%s"
	bannerByIdCacheKey    = "apigw:banner:id:%d"
	bannerActiveCacheKey  = "apigw:banner:active:page:%d:pageSize:%d:search:%s"
	bannerTrashedCacheKey = "apigw:banner:trashed:page:%d:pageSize:%d:search:%s"

	ttlDefault = 5 * time.Minute
)
