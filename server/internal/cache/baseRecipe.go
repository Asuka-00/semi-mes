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
	baseRecipeCachePrefixKey = "baseRecipe:"
	// BaseRecipeExpireTime expire time
	BaseRecipeExpireTime = 5 * time.Minute
)

var _ BaseRecipeCache = (*baseRecipeCache)(nil)

// BaseRecipeCache cache interface
type BaseRecipeCache interface {
	Set(ctx context.Context, id uint64, data *model.BaseRecipe, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.BaseRecipe, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.BaseRecipe, error)
	MultiSet(ctx context.Context, data []*model.BaseRecipe, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// baseRecipeCache define a cache struct
type baseRecipeCache struct {
	cache cache.Cache
}

// NewBaseRecipeCache new a cache
func NewBaseRecipeCache(cacheType *database.CacheType) BaseRecipeCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.BaseRecipe{}
		})
		return &baseRecipeCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.BaseRecipe{}
		})
		return &baseRecipeCache{cache: c}
	}

	return nil // no cache
}

// GetBaseRecipeCacheKey cache key
func (c *baseRecipeCache) GetBaseRecipeCacheKey(id uint64) string {
	return baseRecipeCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *baseRecipeCache) Set(ctx context.Context, id uint64, data *model.BaseRecipe, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetBaseRecipeCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *baseRecipeCache) Get(ctx context.Context, id uint64) (*model.BaseRecipe, error) {
	var data *model.BaseRecipe
	cacheKey := c.GetBaseRecipeCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *baseRecipeCache) MultiSet(ctx context.Context, data []*model.BaseRecipe, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetBaseRecipeCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *baseRecipeCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.BaseRecipe, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetBaseRecipeCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.BaseRecipe)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.BaseRecipe)
	for _, id := range ids {
		val, ok := itemMap[c.GetBaseRecipeCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *baseRecipeCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetBaseRecipeCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *baseRecipeCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetBaseRecipeCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *baseRecipeCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
