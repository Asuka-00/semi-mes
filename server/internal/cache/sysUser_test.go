package cache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/go-dev-frame/sponge/pkg/gotest"
	"github.com/go-dev-frame/sponge/pkg/utils"

	"semi-mes/server/internal/database"
	"semi-mes/server/internal/model"
)

func newSysUserCache() *gotest.Cache {
	record1 := &model.SysUser{}
	record1.ID = 1
	record2 := &model.SysUser{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewSysUserCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_sysUserCache_Set(t *testing.T) {
	c := newSysUserCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.SysUser)
	err := c.ICache.(SysUserCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(SysUserCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_sysUserCache_Get(t *testing.T) {
	c := newSysUserCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.SysUser)
	err := c.ICache.(SysUserCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(SysUserCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(SysUserCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_sysUserCache_MultiGet(t *testing.T) {
	c := newSysUserCache()
	defer c.Close()

	var testData []*model.SysUser
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.SysUser))
	}

	err := c.ICache.(SysUserCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(SysUserCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.SysUser))
	}
}

func Test_sysUserCache_MultiSet(t *testing.T) {
	c := newSysUserCache()
	defer c.Close()

	var testData []*model.SysUser
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.SysUser))
	}

	err := c.ICache.(SysUserCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_sysUserCache_Del(t *testing.T) {
	c := newSysUserCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.SysUser)
	err := c.ICache.(SysUserCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_sysUserCache_SetCacheWithNotFound(t *testing.T) {
	c := newSysUserCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.SysUser)
	err := c.ICache.(SysUserCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(SysUserCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewSysUserCache(t *testing.T) {
	c := NewSysUserCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewSysUserCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewSysUserCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
