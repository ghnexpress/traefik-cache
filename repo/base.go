package repo

import (
	"github.com/ghnexpress/traefik-cache/model"
	"github.com/hoisie/redis"
)

type Repository interface {
	SetExpires(string, int64, model.Cache) error
	Get(string) (*model.Cache, error)
}

type repoManager struct {
	db redis.Client
}

func NewRepoManager(db redis.Client) Repository {
	return &repoManager{db: db}
}
