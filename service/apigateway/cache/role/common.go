package role_cache

import "time"

const (
	roleAllCacheKey      = "apigw:role:all:page:%d:pageSize:%d:search:%s"
	roleByIdCacheKey     = "apigw:role:id:%d"
	roleByUserIdCacheKey = "apigw:role:user_id:%d"
	roleActiveCacheKey   = "apigw:role:active:page:%d:pageSize:%d:search:%s"
	roleTrashedCacheKey  = "apigw:role:trashed:page:%d:pageSize:%d:search:%s"

	ttlDefault = 5 * time.Minute
)
