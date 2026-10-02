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
	sysUserCachePrefixKey = "sysUser:"
	// SysUserExpireTime expire time
	SysUserExpireTime = 5 * time.Minute
)

var _ SysUserCache = (*sysUserCache)(nil)

// SysUserCache cache interface
type SysUserCache interface {
	Set(ctx context.Context, id uint64, data *model.SysUser, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.SysUser, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.SysUser, error)
	MultiSet(ctx context.Context, data []*model.SysUser, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetPlaceholder(ctx context.Context, id uint64) error
	IsPlaceholderErr(err error) bool
}

// sysUserCache define a cache struct
type sysUserCache struct {
	cache cache.Cache
}

// NewSysUserCache new a cache
func NewSysUserCache(cacheType *database.CacheType) SysUserCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.SysUser{}
		})
		return &sysUserCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.SysUser{}
		})
		return &sysUserCache{cache: c}
	}

	return nil // no cache
}

// GetSysUserCacheKey cache key
func (c *sysUserCache) GetSysUserCacheKey(id uint64) string {
	return sysUserCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *sysUserCache) Set(ctx context.Context, id uint64, data *model.SysUser, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetSysUserCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *sysUserCache) Get(ctx context.Context, id uint64) (*model.SysUser, error) {
	var data *model.SysUser
	cacheKey := c.GetSysUserCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *sysUserCache) MultiSet(ctx context.Context, data []*model.SysUser, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetSysUserCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *sysUserCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.SysUser, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetSysUserCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.SysUser)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.SysUser)
	for _, id := range ids {
		val, ok := itemMap[c.GetSysUserCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *sysUserCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetSysUserCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetPlaceholder set placeholder value to cache
func (c *sysUserCache) SetPlaceholder(ctx context.Context, id uint64) error {
	cacheKey := c.GetSysUserCacheKey(id)
	return c.cache.SetCacheWithNotFound(ctx, cacheKey)
}

// IsPlaceholderErr check if cache is placeholder error
func (c *sysUserCache) IsPlaceholderErr(err error) bool {
	return errors.Is(err, cache.ErrPlaceholder)
}
