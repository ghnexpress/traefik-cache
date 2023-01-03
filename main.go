package traefik_cache

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/ghnexpress/traefik-cache/constants"
	"github.com/ghnexpress/traefik-cache/log"
	"github.com/ghnexpress/traefik-cache/model"
	"github.com/ghnexpress/traefik-cache/repo"
	"github.com/ghnexpress/traefik-cache/utils"
	"github.com/pquerna/cachecontrol"
)

var (
	cacheRepo          *repo.Repository
	ignoreHeaderFields = []string{"X-Request-Id", "Postman-Token", "Content-Length"}
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
	config    model.Config
	cacheRepo repo.Repository
}

func New(_ context.Context, next http.Handler, config *model.Config, name string) (http.Handler, error) {
	log.Log("", fmt.Sprintf("config: %+v", *config))

	if cacheRepo == nil {
		client := memcache.New(config.Memcached.Address)
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

func (c *Cache) logError(requestID string, err error) {
	if c.config.Alert.Telegram != nil {
		telegram := c.config.Alert.Telegram

		params := url.Values{}
		params.Add("chat_id", telegram.ChatID)
		params.Add("text", fmt.Sprintf("[%s][cache-middleware-plugin]\nRequestID: %s\n%s", c.config.Env, requestID, err.Error()))
		params.Add("parse_mode", "HTML")

		rs, errGet := http.Get(fmt.Sprintf("https://api.telegram.org/%s/sendMessage?%s", telegram.Token, params.Encode()))
		if errGet != nil {
			log.Log(requestID, errGet.Error())
		}

		if rs.StatusCode != 200 {
			body, errRead := ioutil.ReadAll(rs.Body)
			if errRead != nil {
				log.Log(requestID, errRead.Error())
			}

			rs.Body.Close()
			log.Log(requestID, string(body))
		}
	}

	log.Log(requestID, err.Error())
}

func (c *Cache) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	requestID := req.Header.Get(X_REQUEST_ID_HEADER)

	key, err := c.key(req)
	if err != nil {
		c.logError(requestID, fmt.Errorf("Build key memcached error: %v", err))

		rw.Header().Set(CACHE_HEADER, string(constants.ErrorCacheStatus))

		c.next.ServeHTTP(rw, req)

		return
	}

	value, err := c.cacheRepo.Get(key)
	if err != nil {
		c.logError(requestID, err)

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
			c.logError(requestID, fmt.Errorf("Write data from cache to response body error: %v", err))

			if err := c.cacheRepo.Delete(key); err != nil {
				c.logError(requestID, err)
			}
		}
		return
	}

	rw.Header().Set(CACHE_HEADER, string(constants.MissCacheStatus))

	r := &ResponseWriter{ResponseWriter: rw}

	c.next.ServeHTTP(r, req)

	if expiredTime, ok := c.cacheable(req, rw, r.status); ok {
		err = c.cacheRepo.SetExpires(key, expiredTime, model.Cache{
			Status:  r.status,
			Headers: r.Header(),
			Body:    r.body,
		})

		if err != nil {
			c.logError(requestID, err)
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
