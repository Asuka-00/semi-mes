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
	baseProductCachePrefixKey = "baseProduct:"
	// BaseProductExpireTime expire time
	BaseProductExpireTime = 5 * time.Minute
)

var _ BaseProductCache = (*baseProductCache)(nil)

// BaseProductCache cache interface
type BaseProductCache interface {
	Set(ctx context.Context, id uint64, data *model.BaseProduct, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.BaseProduct, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.BaseProduct, error)
	MultiSet(ctx context.Context, data []*model.BaseProduct, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// baseProductCache define a cache struct
type baseProductCache struct {
	cache cache.Cache
}

// NewBaseProductCache new a cache
func NewBaseProductCache(cacheType *database.CacheType) BaseProductCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.BaseProduct{}
		})
		return &baseProductCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.BaseProduct{}
		})
		return &baseProductCache{cache: c}
	}

	return nil // no cache
}

// GetBaseProductCacheKey cache key
func (c *baseProductCache) GetBaseProductCacheKey(id uint64) string {
	return baseProductCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *baseProductCache) Set(ctx context.Context, id uint64, data *model.BaseProduct, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetBaseProductCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *baseProductCache) Get(ctx context.Context, id uint64) (*model.BaseProduct, error) {
	var data *model.BaseProduct
	cacheKey := c.GetBaseProductCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *baseProductCache) MultiSet(ctx context.Context, data []*model.BaseProduct, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetBaseProductCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *baseProductCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.BaseProduct, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetBaseProductCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.BaseProduct)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.BaseProduct)
	for _, id := range ids {
		val, ok := itemMap[c.GetBaseProductCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *baseProductCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetBaseProductCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *baseProductCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetBaseProductCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *baseProductCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
