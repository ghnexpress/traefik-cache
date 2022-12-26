package traefik_cache

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/ghnexpress/traefik-cache/constants"
	"github.com/ghnexpress/traefik-cache/log"
	"github.com/ghnexpress/traefik-cache/model"
	"github.com/ghnexpress/traefik-cache/repo"
	"github.com/pquerna/cachecontrol"

	"github.com/bradfitz/gomemcache/memcache"
)

var cacheRepo *repo.Repository

const cacheHeader = "Cache-Status"

func CreateConfig() *model.Config {
	return &model.Config{
		AddCacheStatus: true,
	}
}

type Cache struct {
	name      string
	next      http.Handler
	config    model.Config
	cacheRepo repo.Repository
}

func New(_ context.Context, next http.Handler, config *model.Config, name string) (http.Handler, error) {
	if cacheRepo == nil {
		log.Log("2")
		os.Stdout.WriteString("text")
		client := memcache.New(config.Memcached.Address)
		log.Log("3")
		repoManager := repo.NewRepoManager(*client)
		cacheRepo = &repoManager
	}

	return &Cache{
		name:      name,
		next:      next,
		config:    *config,
		cacheRepo: *cacheRepo,
	}, nil
}

func cacheKey(r *http.Request) string {
	return r.Method + r.Host + r.URL.Path
}

func (c *Cache) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	key := cacheKey(req)
	cs := constants.MissCacheStatus

	value, err := c.cacheRepo.Get(key)
	if err != nil {
		cs = constants.ErrorCacheStatus
		log.Log(err.Error())
	}

	if value != nil {
		//lodash
		for key, vals := range value.Headers {
			for _, val := range vals {
				rw.Header().Add(key, val)
			}
		}

		rw.Header().Set(cacheHeader, string(constants.HitCacheStatus))
		rw.WriteHeader(value.Status)
		_, err = rw.Write(value.Body)
		return
	}

	rw.Header().Set(cacheHeader, string(cs))

	r := &ResponseWriter{ResponseWriter: rw}

	c.next.ServeHTTP(r, req)

	if expiredTime, ok := c.cacheable(req, rw, r.status); ok {
		err = c.cacheRepo.SetExpires(key, int32(expiredTime), model.Cache{
			Status:  r.status,
			Headers: r.Header(),
			Body:    r.body,
		})

		if err != nil {
			log.Log(err.Error())
		}
	}
}

func (c *Cache) cacheable(req *http.Request, rw http.ResponseWriter, status int) (int64, bool) {
	reasons, _, err := cachecontrol.CachableResponseWriter(req, status, rw, cachecontrol.Options{})

	if err != nil || len(reasons) > 0 {
		return 0, false
	}

	return time.Now().Unix() + 100, true
}
