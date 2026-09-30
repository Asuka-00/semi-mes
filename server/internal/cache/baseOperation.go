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
	baseOperationCachePrefixKey = "baseOperation:"
	// BaseOperationExpireTime expire time
	BaseOperationExpireTime = 5 * time.Minute
)

var _ BaseOperationCache = (*baseOperationCache)(nil)

// BaseOperationCache cache interface
type BaseOperationCache interface {
	Set(ctx context.Context, id uint64, data *model.BaseOperation, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.BaseOperation, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.BaseOperation, error)
	MultiSet(ctx context.Context, data []*model.BaseOperation, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// baseOperationCache define a cache struct
type baseOperationCache struct {
	cache cache.Cache
}

// NewBaseOperationCache new a cache
func NewBaseOperationCache(cacheType *database.CacheType) BaseOperationCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.BaseOperation{}
		})
		return &baseOperationCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.BaseOperation{}
		})
		return &baseOperationCache{cache: c}
	}

	return nil // no cache
}

// GetBaseOperationCacheKey cache key
func (c *baseOperationCache) GetBaseOperationCacheKey(id uint64) string {
	return baseOperationCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *baseOperationCache) Set(ctx context.Context, id uint64, data *model.BaseOperation, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetBaseOperationCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *baseOperationCache) Get(ctx context.Context, id uint64) (*model.BaseOperation, error) {
	var data *model.BaseOperation
	cacheKey := c.GetBaseOperationCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *baseOperationCache) MultiSet(ctx context.Context, data []*model.BaseOperation, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetBaseOperationCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *baseOperationCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.BaseOperation, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetBaseOperationCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.BaseOperation)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.BaseOperation)
	for _, id := range ids {
		val, ok := itemMap[c.GetBaseOperationCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *baseOperationCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetBaseOperationCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *baseOperationCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetBaseOperationCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *baseOperationCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
