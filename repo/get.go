package repo

import (
	"encoding/json"
	"fmt"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/ghnexpress/traefik-cache/model"
)

func (r *repoManager) Get(key string) (*model.Cache, error) {
	item, err := r.db.Get(key)

	if err != nil {
		if err == memcache.ErrCacheMiss {
			return nil, nil
		}
		return nil, fmt.Errorf("Get data from memcached error: %v", err)
	}

	if item == nil || len(item.Value) == 0 {
		return nil, nil
	}

	var d model.Cache
	if err = json.Unmarshal(item.Value, &d); err != nil {
		return nil, fmt.Errorf("Unmarshal cache data error: %v", err)
	}

	return &d, nil
}
