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

func newSysUserRoleCache() *gotest.Cache {
	record1 := &model.SysUserRole{}
	record1.ID = 1
	record2 := &model.SysUserRole{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewSysUserRoleCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_sysUserRoleCache_Set(t *testing.T) {
	c := newSysUserRoleCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.SysUserRole)
	err := c.ICache.(SysUserRoleCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(SysUserRoleCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_sysUserRoleCache_Get(t *testing.T) {
	c := newSysUserRoleCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.SysUserRole)
	err := c.ICache.(SysUserRoleCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(SysUserRoleCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(SysUserRoleCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_sysUserRoleCache_MultiGet(t *testing.T) {
	c := newSysUserRoleCache()
	defer c.Close()

	var testData []*model.SysUserRole
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.SysUserRole))
	}

	err := c.ICache.(SysUserRoleCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(SysUserRoleCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.SysUserRole))
	}
}

func Test_sysUserRoleCache_MultiSet(t *testing.T) {
	c := newSysUserRoleCache()
	defer c.Close()

	var testData []*model.SysUserRole
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.SysUserRole))
	}

	err := c.ICache.(SysUserRoleCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_sysUserRoleCache_Del(t *testing.T) {
	c := newSysUserRoleCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.SysUserRole)
	err := c.ICache.(SysUserRoleCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_sysUserRoleCache_SetCacheWithNotFound(t *testing.T) {
	c := newSysUserRoleCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.SysUserRole)
	err := c.ICache.(SysUserRoleCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(SysUserRoleCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewSysUserRoleCache(t *testing.T) {
	c := NewSysUserRoleCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewSysUserRoleCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewSysUserRoleCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
