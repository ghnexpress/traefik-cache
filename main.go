package traefik_cache

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/ghnexpress/traefik-cache/constants"
	"github.com/ghnexpress/traefik-cache/log"
	"github.com/ghnexpress/traefik-cache/model"
	"github.com/ghnexpress/traefik-cache/repo"
	"github.com/ghnexpress/traefik-cache/utils"
	"github.com/pquerna/cachecontrol"
)

var (
	cacheRepo           repo.Repository
	onceMemcached       = make(map[string]*sync.Once)
	onceMemcachedMutex  = sync.RWMutex{}
	defaultForceExpired = 60 * 60
	ignoreHeaderFields  = []string{"X-Request-Id", "Postman-Token", "Content-Length"}
)

const (
	X_REQUEST_ID_HEADER = "X-Request-Id"
	CACHE_HEADER        = "cache-status"
	MASTER_ENV          = "master"
	DEV_ENV             = "dev"
)

func CreateConfig() *model.Config {
	return &model.Config{
		Memcached: model.MemcachedConfig{},
		HashKey:   model.HashKey{Method: model.Enable{Enable: true}},
		Env:       MASTER_ENV,
	}
}

type Cache struct {
	name      string
	next      http.Handler
	log       log.Log
	config    model.Config
	cacheRepo repo.Repository
}

func New(_ context.Context, next http.Handler, config *model.Config, name string) (http.Handler, error) {
	log := log.New(config.Env, config.Alert.Telegram)
	log.ConsoleLog("config", config)

	onceMemcachedMutex.Lock()
	onceKey := fmt.Sprintf("%+v", config.Memcached)
	if onceMemcached[onceKey] == nil {
		onceMemcached[onceKey] = &sync.Once{}
	}

	onceMemcached[onceKey].Do(func() {
		cacheRepo = repo.NewRepoManager(config.Memcached)
	})
	onceMemcachedMutex.Unlock()

	return &Cache{
		name:      name,
		next:      next,
		log:       log,
		config:    *config,
		cacheRepo: cacheRepo,
	}, nil
}

func (c *Cache) key(r *http.Request) (string, error) {
	hashKey := c.config.HashKey

	hMethod := ""
	if hashKey.Method.Enable {
		hMethod = r.Method
	}

	hHeader := ""
	if hashKey.Header.Enable && r.Header != nil {
		h := r.Header.Clone()

		if hashKey.Header.Fields != "" {
			rawHeader := ""
			headerFields := strings.Split(hashKey.Header.Fields, ",")
			for _, field := range headerFields {
				rawHeader = fmt.Sprintf("%s|%s", rawHeader, h.Get(field))
			}

			hHeader = utils.GetMD5Hash([]byte(rawHeader))
		} else {
			if hashKey.Header.IgnoreFields != "" {
				ignoreHeaderFields = strings.Split(hashKey.Header.IgnoreFields, ",")
			}

			for _, field := range ignoreHeaderFields {
				h.Del(field)
			}

			hHeader = utils.GetMD5Hash([]byte(fmt.Sprintf("%+v", h)))
		}
	}

	hBody := ""
	if hashKey.Body.Enable && r.Body != nil {
		bodyBytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return "", err
		}

		r.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
		hBody = utils.GetMD5Hash(bodyBytes)
	}

	return utils.GetMD5Hash([]byte(fmt.Sprintf("%s%s|%s|%s|%s", r.Host, r.URL.String(), hMethod, hHeader, hBody))), nil
}

func (c *Cache) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	requestID := req.Header.Get(X_REQUEST_ID_HEADER)

	key, err := c.key(req)
	if err != nil {
		c.log.TelegramLog(requestID, fmt.Errorf("Build key memcached error: %v", err))

		rw.Header().Set(CACHE_HEADER, string(constants.ErrorCacheStatus))

		c.next.ServeHTTP(rw, req)

		return
	}

	value, err := c.cacheRepo.Get(key)
	if err != nil {
		c.log.TelegramLog(requestID, err)

		rw.Header().Set(CACHE_HEADER, string(constants.ErrorCacheStatus))

		c.next.ServeHTTP(rw, req)

		return
	}

	if value != nil {
		for key, vals := range value.Headers {
			for _, val := range vals {
				rw.Header().Add(key, val)
			}
		}

		rw.Header().Set(CACHE_HEADER, string(constants.HitCacheStatus))
		if c.config.Env == DEV_ENV {
			rw.Header().Set("debug-cache-traefik", fmt.Sprintf("time: %s, key: %s", time.Now().Format(time.RFC3339), key))
		}

		rw.WriteHeader(value.Status)
		if _, err := rw.Write(value.Body); err != nil {
			c.log.TelegramLog(requestID, fmt.Errorf("Write data from cache to response body error: %v", err))

			if err := c.cacheRepo.Delete(key); err != nil {
				c.log.TelegramLog(requestID, err)
			}
		}
		return
	}

	rw.Header().Set(CACHE_HEADER, string(constants.MissCacheStatus))
	checkCompress := rw.Header().Get("Vary")

	r := &ResponseWriter{ResponseWriter: rw}

	c.next.ServeHTTP(r, req)

	force := c.config.ForceCache
	ok := force.Enable
	expiredTime := time.Now().Add(time.Second * time.Duration(force.ExpiredTime))
	if ok && force.ExpiredTime <= 0 {
		expiredTime = time.Now().Add(time.Second * time.Duration(defaultForceExpired))
	}
	if !ok {
		expiredTime, ok = c.cacheable(req, rw, r.status)
	}

	if ok {
		// Router --> Compress Middleware --> Cache Middleware --> Service
		if checkCompress != "" {
			r.Header().Del("Content-Encoding")
			r.Header().Del("Vary")
		}

		err = c.cacheRepo.SetExpires(key, expiredTime, model.Cache{
			Status:  r.status,
			Headers: r.Header(),
			Body:    r.body,
		})

		if err != nil {
			c.log.TelegramLog(requestID, err)
		}
	}
}

func (c *Cache) cacheable(req *http.Request, rw http.ResponseWriter, status int) (time.Time, bool) {
	reasons, expiredTime, err := cachecontrol.CachableResponseWriter(req, status, rw, cachecontrol.Options{})

	if err != nil || len(reasons) > 0 || expiredTime.Before(time.Now()) {
		return time.Time{}, false
	}

	return expiredTime, true
}
