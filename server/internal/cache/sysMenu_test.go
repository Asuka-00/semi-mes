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

func newSysMenuCache() *gotest.Cache {
	record1 := &model.SysMenu{}
	record1.ID = 1
	record2 := &model.SysMenu{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewSysMenuCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_sysMenuCache_Set(t *testing.T) {
	c := newSysMenuCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.SysMenu)
	err := c.ICache.(SysMenuCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(SysMenuCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_sysMenuCache_Get(t *testing.T) {
	c := newSysMenuCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.SysMenu)
	err := c.ICache.(SysMenuCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(SysMenuCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(SysMenuCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_sysMenuCache_MultiGet(t *testing.T) {
	c := newSysMenuCache()
	defer c.Close()

	var testData []*model.SysMenu
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.SysMenu))
	}

	err := c.ICache.(SysMenuCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(SysMenuCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.SysMenu))
	}
}

func Test_sysMenuCache_MultiSet(t *testing.T) {
	c := newSysMenuCache()
	defer c.Close()

	var testData []*model.SysMenu
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.SysMenu))
	}

	err := c.ICache.(SysMenuCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_sysMenuCache_Del(t *testing.T) {
	c := newSysMenuCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.SysMenu)
	err := c.ICache.(SysMenuCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_sysMenuCache_SetCacheWithNotFound(t *testing.T) {
	c := newSysMenuCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.SysMenu)
	err := c.ICache.(SysMenuCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(SysMenuCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewSysMenuCache(t *testing.T) {
	c := NewSysMenuCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewSysMenuCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewSysMenuCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
