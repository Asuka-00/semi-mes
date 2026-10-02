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
	sysRoleCachePrefixKey = "sysRole:"
	// SysRoleExpireTime expire time
	SysRoleExpireTime = 5 * time.Minute
)

var _ SysRoleCache = (*sysRoleCache)(nil)

// SysRoleCache cache interface
type SysRoleCache interface {
	Set(ctx context.Context, id uint64, data *model.SysRole, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.SysRole, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.SysRole, error)
	MultiSet(ctx context.Context, data []*model.SysRole, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// sysRoleCache define a cache struct
type sysRoleCache struct {
	cache cache.Cache
}

// NewSysRoleCache new a cache
func NewSysRoleCache(cacheType *database.CacheType) SysRoleCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.SysRole{}
		})
		return &sysRoleCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.SysRole{}
		})
		return &sysRoleCache{cache: c}
	}

	return nil // no cache
}

// GetSysRoleCacheKey cache key
func (c *sysRoleCache) GetSysRoleCacheKey(id uint64) string {
	return sysRoleCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *sysRoleCache) Set(ctx context.Context, id uint64, data *model.SysRole, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetSysRoleCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *sysRoleCache) Get(ctx context.Context, id uint64) (*model.SysRole, error) {
	var data *model.SysRole
	cacheKey := c.GetSysRoleCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *sysRoleCache) MultiSet(ctx context.Context, data []*model.SysRole, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetSysRoleCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *sysRoleCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.SysRole, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetSysRoleCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.SysRole)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.SysRole)
	for _, id := range ids {
		val, ok := itemMap[c.GetSysRoleCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *sysRoleCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetSysRoleCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *sysRoleCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetSysRoleCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *sysRoleCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
