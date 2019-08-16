package httpclient

import (
	"net/http"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/gregjones/httpcache"
	redis_cache "github.com/gregjones/httpcache/redis"
)

func NewClient(conn redis.Conn) *http.Client {
	etagCache := redis_cache.NewWithClient(conn)
	tr := httpcache.NewTransport(etagCache)
	return &http.Client{
		Transport: tr,
		Timeout:   10 * time.Second,
	}
}
