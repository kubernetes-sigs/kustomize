package httpclient

import (
	"net/http"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/gregjones/httpcache"
	rediscache "github.com/gregjones/httpcache/redis"
)

func FromCache(header http.Header) bool {
	return header.Get(httpcache.XFromCache) != ""
}

func NewClient(conn redis.Conn) *http.Client {
	etagCache := rediscache.NewWithClient(conn)
	tr := httpcache.NewTransport(etagCache)
	return &http.Client{
		Transport: tr,
		Timeout:   10 * time.Second,
	}
}
