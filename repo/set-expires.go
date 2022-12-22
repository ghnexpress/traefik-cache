package repo

import (
	"encoding/json"
	"fmt"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/ghnexpress/traefik-cache/model"
)

func (r *repoManager) SetExpires(key string, time int32, data model.Cache) error {
	b, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("Marshal cache data error: %v", err)
	}

	err = r.db.Set(&memcache.Item{
		Key:        key,
		Value:      b,
		Expiration: time})

	if err != nil {
		return fmt.Errorf("Set data to memcached error: %v", err)
	}

	return nil
}
