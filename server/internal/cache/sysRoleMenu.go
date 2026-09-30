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
	sysRoleMenuCachePrefixKey = "sysRoleMenu:"
	// SysRoleMenuExpireTime expire time
	SysRoleMenuExpireTime = 5 * time.Minute
)

var _ SysRoleMenuCache = (*sysRoleMenuCache)(nil)

// SysRoleMenuCache cache interface
type SysRoleMenuCache interface {
	Set(ctx context.Context, id uint64, data *model.SysRoleMenu, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.SysRoleMenu, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.SysRoleMenu, error)
	MultiSet(ctx context.Context, data []*model.SysRoleMenu, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// sysRoleMenuCache define a cache struct
type sysRoleMenuCache struct {
	cache cache.Cache
}

// NewSysRoleMenuCache new a cache
func NewSysRoleMenuCache(cacheType *database.CacheType) SysRoleMenuCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.SysRoleMenu{}
		})
		return &sysRoleMenuCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.SysRoleMenu{}
		})
		return &sysRoleMenuCache{cache: c}
	}

	return nil // no cache
}

// GetSysRoleMenuCacheKey cache key
func (c *sysRoleMenuCache) GetSysRoleMenuCacheKey(id uint64) string {
	return sysRoleMenuCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *sysRoleMenuCache) Set(ctx context.Context, id uint64, data *model.SysRoleMenu, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetSysRoleMenuCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *sysRoleMenuCache) Get(ctx context.Context, id uint64) (*model.SysRoleMenu, error) {
	var data *model.SysRoleMenu
	cacheKey := c.GetSysRoleMenuCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *sysRoleMenuCache) MultiSet(ctx context.Context, data []*model.SysRoleMenu, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetSysRoleMenuCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *sysRoleMenuCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.SysRoleMenu, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetSysRoleMenuCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.SysRoleMenu)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.SysRoleMenu)
	for _, id := range ids {
		val, ok := itemMap[c.GetSysRoleMenuCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *sysRoleMenuCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetSysRoleMenuCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *sysRoleMenuCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetSysRoleMenuCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *sysRoleMenuCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
