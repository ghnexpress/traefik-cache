package repo

import (
	"fmt"

	"github.com/bradfitz/gomemcache/memcache"
)

func (r *repoManager) Delete(key string) error {
	err := r.db.Delete(key)

	if err != nil {
		if err == memcache.ErrCacheMiss {
			return nil
		}
		return fmt.Errorf("Delete data from memcached error: %v", err)
	}

	return nil
}
