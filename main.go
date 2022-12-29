package traefik_cache

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/ghnexpress/traefik-cache/constants"
	"github.com/ghnexpress/traefik-cache/log"
	"github.com/ghnexpress/traefik-cache/model"
	"github.com/ghnexpress/traefik-cache/repo"
	"github.com/ghnexpress/traefik-cache/utils"
	"github.com/pquerna/cachecontrol"

	"github.com/bradfitz/gomemcache/memcache"
)

var (
	cacheRepo          *repo.Repository
	ignoreHeaderFields = []string{"X-Request-Id", "Postman-Token", "Content-Length"}
)

const cacheHeader = "Cache-Status"

func CreateConfig() *model.Config {
	return &model.Config{
		Memcached: model.MemcachedConfig{},
		HashKey: model.HashKey{
			Method: model.Enable{
				Enable: true,
			},
		},
	}
}

type Cache struct {
	name      string
	next      http.Handler
	config    model.Config
	cacheRepo repo.Repository
}

func New(_ context.Context, next http.Handler, config *model.Config, name string) (http.Handler, error) {
	log.Log(fmt.Sprintf("config: %+v", *config))

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

func (c *Cache) key(r *http.Request) string {
	hashKey := c.config.HashKey

	hMethod := ""
	if hashKey.Method.Enable {
		hMethod = r.Method
	}

	hHeader := ""
	if hashKey.Header.Enable && r.Header != nil {
		h := r.Header.Clone()

		if hashKey.Header.IgnoreFields != "" {
			ignoreHeaderFields = strings.Split(hashKey.Header.IgnoreFields, ",")
		}

		for _, field := range ignoreHeaderFields {
			h.Del(field)
		}

		hHeader = utils.GetMD5Hash([]byte(fmt.Sprintf("%+v", h)))
		log.Log(fmt.Sprintf("header: %s", fmt.Sprintf("%+v", h)))
	}

	hBody := ""
	if hashKey.Body.Enable && r.Body != nil {
		bodyBytes, _ := ioutil.ReadAll(r.Body)
		r.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

		hBody = utils.GetMD5Hash(bodyBytes)
		log.Log(fmt.Sprintf("body: %s", bodyBytes))
	}

	key := fmt.Sprintf("%s%s|%s|%s|%s", r.Host, r.URL.String(), hMethod, hHeader, hBody)

	return key
}

func (c *Cache) logError(err error) {
	if c.config.Alert.Telegram != nil {
		telegram := c.config.Alert.Telegram

		params := url.Values{}
		params.Add("chat_id", telegram.ChatID)
		params.Add("text", fmt.Sprintf("[%s][cache-middleware-plugin] \n%s", os.Getenv("APPLICATION_ENV"), err.Error()))
		params.Add("parse_mode", "HTML")

		http.Get(fmt.Sprintf("https://api.telegram.org/%s/sendMessage?%s", telegram.Token, params.Encode()))
	}

	log.Log(err.Error())
}

func (c *Cache) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	key := c.key(req)
	cs := constants.MissCacheStatus

	value, err := c.cacheRepo.Get(key)
	if err != nil {
		cs = constants.ErrorCacheStatus
		c.logError(err)
	}

	if value != nil {
		for key, vals := range value.Headers {
			for _, val := range vals {
				rw.Header().Add(key, val)
			}
		}

		rw.Header().Set(cacheHeader, string(constants.HitCacheStatus))
		rw.Header().Set("debug", fmt.Sprintf("time: %s", time.Now().String()))

		rw.WriteHeader(value.Status)
		_, err = rw.Write(value.Body)
		return
	}

	rw.Header().Set(cacheHeader, string(cs))

	r := &ResponseWriter{ResponseWriter: rw}

	c.next.ServeHTTP(r, req)

	if expiredTime, ok := c.cacheable(req, rw, r.status); ok {
		err = c.cacheRepo.SetExpires(key, expiredTime, model.Cache{
			Status:  r.status,
			Headers: r.Header(),
			Body:    r.body,
		})

		if err != nil {
			c.logError(err)
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
