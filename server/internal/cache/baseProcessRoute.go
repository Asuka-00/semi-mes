package cache

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/go-dev-frame/sponge/pkg/cache"
	"github.com/go-dev-frame/sponge/pkg/encoding"
	"github.com/go-dev-frame/sponge/pkg/utils"

	"semi-mes/server/internal/database"
	"semi-mes/server/internal/model"
)

const (
	// cache prefix key, must end with a colon
	baseProcessRouteCachePrefixKey = "baseProcessRoute:"
	// BaseProcessRouteExpireTime expire time
	BaseProcessRouteExpireTime = 5 * time.Minute
)

var _ BaseProcessRouteCache = (*baseProcessRouteCache)(nil)

// BaseProcessRouteCache cache interface
type BaseProcessRouteCache interface {
	Set(ctx context.Context, id uint64, data *model.BaseProcessRoute, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.BaseProcessRoute, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.BaseProcessRoute, error)
	MultiSet(ctx context.Context, data []*model.BaseProcessRoute, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// baseProcessRouteCache define a cache struct
type baseProcessRouteCache struct {
	cache cache.Cache
}

// NewBaseProcessRouteCache new a cache
func NewBaseProcessRouteCache(cacheType *database.CacheType) BaseProcessRouteCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.BaseProcessRoute{}
		})
		return &baseProcessRouteCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.BaseProcessRoute{}
		})
		return &baseProcessRouteCache{cache: c}
	}

	return nil // no cache
}

// GetBaseProcessRouteCacheKey cache key
func (c *baseProcessRouteCache) GetBaseProcessRouteCacheKey(id uint64) string {
	return baseProcessRouteCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *baseProcessRouteCache) Set(ctx context.Context, id uint64, data *model.BaseProcessRoute, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetBaseProcessRouteCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *baseProcessRouteCache) Get(ctx context.Context, id uint64) (*model.BaseProcessRoute, error) {
	var data *model.BaseProcessRoute
	cacheKey := c.GetBaseProcessRouteCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *baseProcessRouteCache) MultiSet(ctx context.Context, data []*model.BaseProcessRoute, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetBaseProcessRouteCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *baseProcessRouteCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.BaseProcessRoute, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetBaseProcessRouteCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.BaseProcessRoute)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.BaseProcessRoute)
	for _, id := range ids {
		val, ok := itemMap[c.GetBaseProcessRouteCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *baseProcessRouteCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetBaseProcessRouteCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *baseProcessRouteCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetBaseProcessRouteCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *baseProcessRouteCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
