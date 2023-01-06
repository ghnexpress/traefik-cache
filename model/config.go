package model

type MemcachedConfig struct {
	Address string `json:"address,omitempty"`
}

type Enable struct {
	Enable bool `json:"enable,omitempty"`
}

type HeaderHashKey struct {
	Enable       bool   `json:"enable,omitempty"`
	Fields       string `json:"fields,omitempty"`
	IgnoreFields string `json:"ignoreFields,omitempty"`
}

type HashKey struct {
	Method Enable        `json:"method,omitempty"`
	Header HeaderHashKey `json:"header,omitempty"`
	Body   Enable        `json:"body,omitempty"`
}

type TelegramConfig struct {
	ChatID string `json:"chatId,omitempty"`
	Token  string `json:"token,omitempty"`
}

type AlertConfig struct {
	Telegram *TelegramConfig `json:"telegram,omitempty"`
}

type ForceCache struct {
	Enable      bool `json:"enable,omitempty"`
	ExpiredTime int  `json:"expiredTime,omitempty"`
}

type Config struct {
	Memcached  MemcachedConfig `json:"memcached,omitempty"`
	HashKey    HashKey         `json:"hashkey,omitempty"`
	Alert      AlertConfig     `json:"alert,omitempty"`
	ForceCache ForceCache      `json:"forceCache,omitempty"`
	Env        string          `json:"env,omitempty"`
}
