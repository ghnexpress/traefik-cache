package model

type MemcachedConfig struct {
	Address string `json:"address,omitempty"`
}

type HashKey struct {
	Method bool `json:"method,omitempty"`
	Header bool `json:"header,omitempty"`
	Body   bool `json:"body,omitempty"`
}

type Config struct {
	Memcached MemcachedConfig `json:"memcached,omitempty"`
	HashKey   HashKey         `json:"hashkey,omitempty"`
}
