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

func newBaseFactoryCache() *gotest.Cache {
	record1 := &model.BaseFactory{}
	record1.ID = 1
	record2 := &model.BaseFactory{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewBaseFactoryCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_baseFactoryCache_Set(t *testing.T) {
	c := newBaseFactoryCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.BaseFactory)
	err := c.ICache.(BaseFactoryCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(BaseFactoryCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_baseFactoryCache_Get(t *testing.T) {
	c := newBaseFactoryCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.BaseFactory)
	err := c.ICache.(BaseFactoryCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(BaseFactoryCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(BaseFactoryCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_baseFactoryCache_MultiGet(t *testing.T) {
	c := newBaseFactoryCache()
	defer c.Close()

	var testData []*model.BaseFactory
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.BaseFactory))
	}

	err := c.ICache.(BaseFactoryCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(BaseFactoryCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.BaseFactory))
	}
}

func Test_baseFactoryCache_MultiSet(t *testing.T) {
	c := newBaseFactoryCache()
	defer c.Close()

	var testData []*model.BaseFactory
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.BaseFactory))
	}

	err := c.ICache.(BaseFactoryCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_baseFactoryCache_Del(t *testing.T) {
	c := newBaseFactoryCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.BaseFactory)
	err := c.ICache.(BaseFactoryCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_baseFactoryCache_SetCacheWithNotFound(t *testing.T) {
	c := newBaseFactoryCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.BaseFactory)
	err := c.ICache.(BaseFactoryCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(BaseFactoryCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewBaseFactoryCache(t *testing.T) {
	c := NewBaseFactoryCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewBaseFactoryCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewBaseFactoryCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
