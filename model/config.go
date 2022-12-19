package model

type RedisConfig struct {
	Address  string `json:"address,omitempty"`
	Password string `json:"password,omitempty"`
}

type Config struct {
	MaxExpiry      int         `json:"maxExpiry" yaml:"maxExpiry" toml:"maxExpiry"`
	AddCacheStatus bool        `json:"addCacheStatus" yaml:"addCacheStatus" toml:"addCacheStatus"`
	Redis          RedisConfig `json:"redis,omitempty"`
}
