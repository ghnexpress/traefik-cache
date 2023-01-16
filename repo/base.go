package repo

import (
	"fmt"
	"os"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/ghnexpress/traefik-cache/model"
)

type Repository interface {
	SetExpires(string, time.Time, model.Cache) error
	Get(string) (*model.Cache, error)
	Delete(string) error
}

type repoManager struct {
	db memcache.Client
}

func NewRepoManager(cfg model.MemcachedConfig) Repository {
	client := memcache.New(cfg.Address)

	if cfg.MaxIdleConnection > 0 {
		client.MaxIdleConns = cfg.MaxIdleConnection
	}

	if cfg.Timeout > 0 {
		client.Timeout = time.Duration(cfg.Timeout) * time.Second
	}

	os.Stdout.WriteString(fmt.Sprintf("[cache-middleware-plugin] [memcached] Memcached connected, config: %+v\n", client))

	return &repoManager{db: *client}
}
