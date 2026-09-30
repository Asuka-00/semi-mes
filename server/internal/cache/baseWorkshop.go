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
	baseWorkshopCachePrefixKey = "baseWorkshop:"
	// BaseWorkshopExpireTime expire time
	BaseWorkshopExpireTime = 5 * time.Minute
)

var _ BaseWorkshopCache = (*baseWorkshopCache)(nil)

// BaseWorkshopCache cache interface
type BaseWorkshopCache interface {
	Set(ctx context.Context, id uint64, data *model.BaseWorkshop, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.BaseWorkshop, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.BaseWorkshop, error)
	MultiSet(ctx context.Context, data []*model.BaseWorkshop, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// baseWorkshopCache define a cache struct
type baseWorkshopCache struct {
	cache cache.Cache
}

// NewBaseWorkshopCache new a cache
func NewBaseWorkshopCache(cacheType *database.CacheType) BaseWorkshopCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.BaseWorkshop{}
		})
		return &baseWorkshopCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.BaseWorkshop{}
		})
		return &baseWorkshopCache{cache: c}
	}

	return nil // no cache
}

// GetBaseWorkshopCacheKey cache key
func (c *baseWorkshopCache) GetBaseWorkshopCacheKey(id uint64) string {
	return baseWorkshopCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *baseWorkshopCache) Set(ctx context.Context, id uint64, data *model.BaseWorkshop, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetBaseWorkshopCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *baseWorkshopCache) Get(ctx context.Context, id uint64) (*model.BaseWorkshop, error) {
	var data *model.BaseWorkshop
	cacheKey := c.GetBaseWorkshopCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *baseWorkshopCache) MultiSet(ctx context.Context, data []*model.BaseWorkshop, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetBaseWorkshopCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *baseWorkshopCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.BaseWorkshop, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetBaseWorkshopCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.BaseWorkshop)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.BaseWorkshop)
	for _, id := range ids {
		val, ok := itemMap[c.GetBaseWorkshopCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *baseWorkshopCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetBaseWorkshopCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *baseWorkshopCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetBaseWorkshopCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *baseWorkshopCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
