package constants

type CacheStatus string

const (
	HitCacheStatus   CacheStatus = "hit"
	MissCacheStatus  CacheStatus = "miss"
	ErrorCacheStatus CacheStatus = "error"
)
