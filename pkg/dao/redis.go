package dao

import (
	"fmt"

	"github.com/go-redis/redis"
)

const (
	// cache key
	cachePre      = "surl:"
	cacheIDKey    = cachePre + "id"
	cacheShortKey = cachePre + "short:%s"
	cacheLongKey  = cachePre + "long:%s"

	// cache ttl
	cacheTTL = 0

	// default ID
	defaultID = 10000
)

var _r r

type r struct {
}

func (r) generateID() (int64, error) {
	key := cacheIDKey
	val, err := re.Incr(key).Result()

	// first access
	if val == 1 {
		val = defaultID
		_, err = re.Set(key, val, 0).Result()
		if err != nil {
			return 0, err
		}
	}

	return val, err
}

func (r) getLongURL(idKey string) (string, error) {
	key := fmt.Sprintf(cacheLongKey, idKey)
	val, err := re.Get(key).Result()
	if err == redis.Nil {
		err = nil
	}
	return val, err
}

func (r) getShortURL(idKey string) (string, error) {
	k := fmt.Sprintf(cacheShortKey, idKey)
	v, err := re.Get(k).Result()
	if err == redis.Nil {
		err = nil
	}
	return v, err
}

func (r) setLongURL(idKey string, val string) error {
	key := fmt.Sprintf(cacheLongKey, idKey)
	_, err := re.Set(key, val, cacheTTL).Result()
	return err
}

func (r) setShortURL(idKey string, val string) error {
	key := fmt.Sprintf(cacheShortKey, idKey)
	_, err := re.Set(key, val, cacheTTL).Result()
	return err
}
