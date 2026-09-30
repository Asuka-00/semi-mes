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

func newSysRoleMenuCache() *gotest.Cache {
	record1 := &model.SysRoleMenu{}
	record1.ID = 1
	record2 := &model.SysRoleMenu{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewSysRoleMenuCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_sysRoleMenuCache_Set(t *testing.T) {
	c := newSysRoleMenuCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.SysRoleMenu)
	err := c.ICache.(SysRoleMenuCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(SysRoleMenuCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_sysRoleMenuCache_Get(t *testing.T) {
	c := newSysRoleMenuCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.SysRoleMenu)
	err := c.ICache.(SysRoleMenuCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(SysRoleMenuCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(SysRoleMenuCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_sysRoleMenuCache_MultiGet(t *testing.T) {
	c := newSysRoleMenuCache()
	defer c.Close()

	var testData []*model.SysRoleMenu
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.SysRoleMenu))
	}

	err := c.ICache.(SysRoleMenuCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(SysRoleMenuCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.SysRoleMenu))
	}
}

func Test_sysRoleMenuCache_MultiSet(t *testing.T) {
	c := newSysRoleMenuCache()
	defer c.Close()

	var testData []*model.SysRoleMenu
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.SysRoleMenu))
	}

	err := c.ICache.(SysRoleMenuCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_sysRoleMenuCache_Del(t *testing.T) {
	c := newSysRoleMenuCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.SysRoleMenu)
	err := c.ICache.(SysRoleMenuCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_sysRoleMenuCache_SetCacheWithNotFound(t *testing.T) {
	c := newSysRoleMenuCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.SysRoleMenu)
	err := c.ICache.(SysRoleMenuCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(SysRoleMenuCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewSysRoleMenuCache(t *testing.T) {
	c := NewSysRoleMenuCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewSysRoleMenuCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewSysRoleMenuCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
