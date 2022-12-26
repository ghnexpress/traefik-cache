package repo

import (
	"fmt"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/ghnexpress/traefik-cache/log"
	"github.com/ghnexpress/traefik-cache/model"
)

type Repository interface {
	SetExpires(string, int32, model.Cache) error
	Get(string) (*model.Cache, error)
}

type repoManager struct {
	db memcache.Client
}

func NewRepoManager(db memcache.Client) Repository {
	if err := db.Ping(); err != nil {
		log.Log(fmt.Sprintf("Could not ping to memcached: %v", err))
		return nil
	}

	log.Log("Memcached connected!")

	return &repoManager{db: db}
}
