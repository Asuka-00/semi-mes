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

func newBaseProductCache() *gotest.Cache {
	record1 := &model.BaseProduct{}
	record1.ID = 1
	record2 := &model.BaseProduct{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewBaseProductCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_baseProductCache_Set(t *testing.T) {
	c := newBaseProductCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.BaseProduct)
	err := c.ICache.(BaseProductCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(BaseProductCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_baseProductCache_Get(t *testing.T) {
	c := newBaseProductCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.BaseProduct)
	err := c.ICache.(BaseProductCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(BaseProductCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(BaseProductCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_baseProductCache_MultiGet(t *testing.T) {
	c := newBaseProductCache()
	defer c.Close()

	var testData []*model.BaseProduct
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.BaseProduct))
	}

	err := c.ICache.(BaseProductCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(BaseProductCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.BaseProduct))
	}
}

func Test_baseProductCache_MultiSet(t *testing.T) {
	c := newBaseProductCache()
	defer c.Close()

	var testData []*model.BaseProduct
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.BaseProduct))
	}

	err := c.ICache.(BaseProductCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_baseProductCache_Del(t *testing.T) {
	c := newBaseProductCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.BaseProduct)
	err := c.ICache.(BaseProductCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_baseProductCache_SetCacheWithNotFound(t *testing.T) {
	c := newBaseProductCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.BaseProduct)
	err := c.ICache.(BaseProductCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(BaseProductCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewBaseProductCache(t *testing.T) {
	c := NewBaseProductCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewBaseProductCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewBaseProductCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
