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

func newBaseProcessRouteCache() *gotest.Cache {
	record1 := &model.BaseProcessRoute{}
	record1.ID = 1
	record2 := &model.BaseProcessRoute{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewBaseProcessRouteCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_baseProcessRouteCache_Set(t *testing.T) {
	c := newBaseProcessRouteCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.BaseProcessRoute)
	err := c.ICache.(BaseProcessRouteCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(BaseProcessRouteCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_baseProcessRouteCache_Get(t *testing.T) {
	c := newBaseProcessRouteCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.BaseProcessRoute)
	err := c.ICache.(BaseProcessRouteCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(BaseProcessRouteCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(BaseProcessRouteCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_baseProcessRouteCache_MultiGet(t *testing.T) {
	c := newBaseProcessRouteCache()
	defer c.Close()

	var testData []*model.BaseProcessRoute
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.BaseProcessRoute))
	}

	err := c.ICache.(BaseProcessRouteCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(BaseProcessRouteCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.BaseProcessRoute))
	}
}

func Test_baseProcessRouteCache_MultiSet(t *testing.T) {
	c := newBaseProcessRouteCache()
	defer c.Close()

	var testData []*model.BaseProcessRoute
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.BaseProcessRoute))
	}

	err := c.ICache.(BaseProcessRouteCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_baseProcessRouteCache_Del(t *testing.T) {
	c := newBaseProcessRouteCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.BaseProcessRoute)
	err := c.ICache.(BaseProcessRouteCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_baseProcessRouteCache_SetCacheWithNotFound(t *testing.T) {
	c := newBaseProcessRouteCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.BaseProcessRoute)
	err := c.ICache.(BaseProcessRouteCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(BaseProcessRouteCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewBaseProcessRouteCache(t *testing.T) {
	c := NewBaseProcessRouteCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewBaseProcessRouteCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewBaseProcessRouteCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
