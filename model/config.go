package model

type MemcachedConfig struct {
	Address string `json:"address,omitempty"`
}

type Config struct {
	AddCacheStatus bool            `json:"addCacheStatus" yaml:"addCacheStatus" toml:"addCacheStatus"`
	Memcached      MemcachedConfig `json:"memcached,omitempty"`
}
