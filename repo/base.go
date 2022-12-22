package repo

import (
	"github.com/bradfitz/gomemcache/memcache"
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
	return &repoManager{db: db}
}
