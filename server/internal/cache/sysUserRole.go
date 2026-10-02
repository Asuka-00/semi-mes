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
	sysUserRoleCachePrefixKey = "sysUserRole:"
	// SysUserRoleExpireTime expire time
	SysUserRoleExpireTime = 5 * time.Minute
)

var _ SysUserRoleCache = (*sysUserRoleCache)(nil)

// SysUserRoleCache cache interface
type SysUserRoleCache interface {
	Set(ctx context.Context, id uint64, data *model.SysUserRole, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.SysUserRole, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.SysUserRole, error)
	MultiSet(ctx context.Context, data []*model.SysUserRole, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// sysUserRoleCache define a cache struct
type sysUserRoleCache struct {
	cache cache.Cache
}

// NewSysUserRoleCache new a cache
func NewSysUserRoleCache(cacheType *database.CacheType) SysUserRoleCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.SysUserRole{}
		})
		return &sysUserRoleCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.SysUserRole{}
		})
		return &sysUserRoleCache{cache: c}
	}

	return nil // no cache
}

// GetSysUserRoleCacheKey cache key
func (c *sysUserRoleCache) GetSysUserRoleCacheKey(id uint64) string {
	return sysUserRoleCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *sysUserRoleCache) Set(ctx context.Context, id uint64, data *model.SysUserRole, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetSysUserRoleCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *sysUserRoleCache) Get(ctx context.Context, id uint64) (*model.SysUserRole, error) {
	var data *model.SysUserRole
	cacheKey := c.GetSysUserRoleCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *sysUserRoleCache) MultiSet(ctx context.Context, data []*model.SysUserRole, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetSysUserRoleCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *sysUserRoleCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.SysUserRole, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetSysUserRoleCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.SysUserRole)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.SysUserRole)
	for _, id := range ids {
		val, ok := itemMap[c.GetSysUserRoleCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *sysUserRoleCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetSysUserRoleCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *sysUserRoleCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetSysUserRoleCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *sysUserRoleCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
