package repo

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/ghnexpress/traefik-cache/model"
)

func (r *repoManager) SetExpires(key string, t time.Time, data model.Cache) error {
	b, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("Marshal cache data error: %v", err)
	}

	expiration := int32(t.Unix() - time.Now().Unix())
	if expiration <= 0 {
		return nil
	}

	err = r.db.Set(&memcache.Item{
		Key:        key,
		Value:      b,
		Expiration: expiration,
	})

	if err != nil {
		return fmt.Errorf("Set data to memcached error: %v", err)
	}

	return nil
}
