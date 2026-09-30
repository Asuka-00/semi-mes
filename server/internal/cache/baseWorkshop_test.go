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

func newBaseWorkshopCache() *gotest.Cache {
	record1 := &model.BaseWorkshop{}
	record1.ID = 1
	record2 := &model.BaseWorkshop{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewBaseWorkshopCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_baseWorkshopCache_Set(t *testing.T) {
	c := newBaseWorkshopCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.BaseWorkshop)
	err := c.ICache.(BaseWorkshopCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(BaseWorkshopCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_baseWorkshopCache_Get(t *testing.T) {
	c := newBaseWorkshopCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.BaseWorkshop)
	err := c.ICache.(BaseWorkshopCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(BaseWorkshopCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(BaseWorkshopCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_baseWorkshopCache_MultiGet(t *testing.T) {
	c := newBaseWorkshopCache()
	defer c.Close()

	var testData []*model.BaseWorkshop
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.BaseWorkshop))
	}

	err := c.ICache.(BaseWorkshopCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(BaseWorkshopCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.BaseWorkshop))
	}
}

func Test_baseWorkshopCache_MultiSet(t *testing.T) {
	c := newBaseWorkshopCache()
	defer c.Close()

	var testData []*model.BaseWorkshop
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.BaseWorkshop))
	}

	err := c.ICache.(BaseWorkshopCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_baseWorkshopCache_Del(t *testing.T) {
	c := newBaseWorkshopCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.BaseWorkshop)
	err := c.ICache.(BaseWorkshopCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_baseWorkshopCache_SetCacheWithNotFound(t *testing.T) {
	c := newBaseWorkshopCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.BaseWorkshop)
	err := c.ICache.(BaseWorkshopCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(BaseWorkshopCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewBaseWorkshopCache(t *testing.T) {
	c := NewBaseWorkshopCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewBaseWorkshopCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewBaseWorkshopCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
