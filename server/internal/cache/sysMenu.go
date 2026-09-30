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
	sysMenuCachePrefixKey = "sysMenu:"
	// SysMenuExpireTime expire time
	SysMenuExpireTime = 5 * time.Minute
)

var _ SysMenuCache = (*sysMenuCache)(nil)

// SysMenuCache cache interface
type SysMenuCache interface {
	Set(ctx context.Context, id uint64, data *model.SysMenu, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.SysMenu, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.SysMenu, error)
	MultiSet(ctx context.Context, data []*model.SysMenu, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// sysMenuCache define a cache struct
type sysMenuCache struct {
	cache cache.Cache
}

// NewSysMenuCache new a cache
func NewSysMenuCache(cacheType *database.CacheType) SysMenuCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.SysMenu{}
		})
		return &sysMenuCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.SysMenu{}
		})
		return &sysMenuCache{cache: c}
	}

	return nil // no cache
}

// GetSysMenuCacheKey cache key
func (c *sysMenuCache) GetSysMenuCacheKey(id uint64) string {
	return sysMenuCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *sysMenuCache) Set(ctx context.Context, id uint64, data *model.SysMenu, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetSysMenuCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *sysMenuCache) Get(ctx context.Context, id uint64) (*model.SysMenu, error) {
	var data *model.SysMenu
	cacheKey := c.GetSysMenuCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *sysMenuCache) MultiSet(ctx context.Context, data []*model.SysMenu, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetSysMenuCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *sysMenuCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.SysMenu, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetSysMenuCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.SysMenu)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.SysMenu)
	for _, id := range ids {
		val, ok := itemMap[c.GetSysMenuCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *sysMenuCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetSysMenuCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *sysMenuCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetSysMenuCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *sysMenuCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
