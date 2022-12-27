package model

type MemcachedConfig struct {
	Address string `json:"address,omitempty"`
}

type Enable struct {
	Enable bool `json:"enable,omitempty"`
}

type HeaderHashKey struct {
	Enable       bool   `json:"enable,omitempty"`
	IgnoreFields string `json:"ignoreFields,omitempty"`
}

type HashKey struct {
	Method Enable        `json:"method,omitempty"`
	Header HeaderHashKey `json:"header,omitempty"`
	Body   Enable        `json:"body,omitempty"`
}

type Config struct {
	Memcached MemcachedConfig `json:"memcached,omitempty"`
	HashKey   HashKey         `json:"hashkey,omitempty"`
}
