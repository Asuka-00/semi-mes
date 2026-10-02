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

func newSysRoleCache() *gotest.Cache {
	record1 := &model.SysRole{}
	record1.ID = 1
	record2 := &model.SysRole{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewSysRoleCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_sysRoleCache_Set(t *testing.T) {
	c := newSysRoleCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.SysRole)
	err := c.ICache.(SysRoleCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(SysRoleCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_sysRoleCache_Get(t *testing.T) {
	c := newSysRoleCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.SysRole)
	err := c.ICache.(SysRoleCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(SysRoleCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(SysRoleCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_sysRoleCache_MultiGet(t *testing.T) {
	c := newSysRoleCache()
	defer c.Close()

	var testData []*model.SysRole
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.SysRole))
	}

	err := c.ICache.(SysRoleCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(SysRoleCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.SysRole))
	}
}

func Test_sysRoleCache_MultiSet(t *testing.T) {
	c := newSysRoleCache()
	defer c.Close()

	var testData []*model.SysRole
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.SysRole))
	}

	err := c.ICache.(SysRoleCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_sysRoleCache_Del(t *testing.T) {
	c := newSysRoleCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.SysRole)
	err := c.ICache.(SysRoleCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_sysRoleCache_SetCacheWithNotFound(t *testing.T) {
	c := newSysRoleCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.SysRole)
	err := c.ICache.(SysRoleCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(SysRoleCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewSysRoleCache(t *testing.T) {
	c := NewSysRoleCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewSysRoleCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewSysRoleCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
