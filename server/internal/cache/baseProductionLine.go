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
	baseProductionLineCachePrefixKey = "baseProductionLine:"
	// BaseProductionLineExpireTime expire time
	BaseProductionLineExpireTime = 5 * time.Minute
)

var _ BaseProductionLineCache = (*baseProductionLineCache)(nil)

// BaseProductionLineCache cache interface
type BaseProductionLineCache interface {
	Set(ctx context.Context, id uint64, data *model.BaseProductionLine, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.BaseProductionLine, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.BaseProductionLine, error)
	MultiSet(ctx context.Context, data []*model.BaseProductionLine, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// baseProductionLineCache define a cache struct
type baseProductionLineCache struct {
	cache cache.Cache
}

// NewBaseProductionLineCache new a cache
func NewBaseProductionLineCache(cacheType *database.CacheType) BaseProductionLineCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.BaseProductionLine{}
		})
		return &baseProductionLineCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.BaseProductionLine{}
		})
		return &baseProductionLineCache{cache: c}
	}

	return nil // no cache
}

// GetBaseProductionLineCacheKey cache key
func (c *baseProductionLineCache) GetBaseProductionLineCacheKey(id uint64) string {
	return baseProductionLineCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *baseProductionLineCache) Set(ctx context.Context, id uint64, data *model.BaseProductionLine, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetBaseProductionLineCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *baseProductionLineCache) Get(ctx context.Context, id uint64) (*model.BaseProductionLine, error) {
	var data *model.BaseProductionLine
	cacheKey := c.GetBaseProductionLineCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *baseProductionLineCache) MultiSet(ctx context.Context, data []*model.BaseProductionLine, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetBaseProductionLineCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *baseProductionLineCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.BaseProductionLine, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetBaseProductionLineCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.BaseProductionLine)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.BaseProductionLine)
	for _, id := range ids {
		val, ok := itemMap[c.GetBaseProductionLineCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *baseProductionLineCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetBaseProductionLineCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *baseProductionLineCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetBaseProductionLineCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *baseProductionLineCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
