package user_cache

import "time"

const (
	userAllCacheKey     = "apigw:user:all:page:%d:pageSize:%d:search:%s"
	userByIdCacheKey    = "apigw:user:id:%d"
	userActiveCacheKey  = "apigw:user:active:page:%d:pageSize:%d:search:%s"
	userTrashedCacheKey = "apigw:user:trashed:page:%d:pageSize:%d:search:%s"

	ttlDefault = 5 * time.Minute
)
