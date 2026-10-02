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

func newBaseOperationCache() *gotest.Cache {
	record1 := &model.BaseOperation{}
	record1.ID = 1
	record2 := &model.BaseOperation{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewBaseOperationCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_baseOperationCache_Set(t *testing.T) {
	c := newBaseOperationCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.BaseOperation)
	err := c.ICache.(BaseOperationCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(BaseOperationCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_baseOperationCache_Get(t *testing.T) {
	c := newBaseOperationCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.BaseOperation)
	err := c.ICache.(BaseOperationCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(BaseOperationCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(BaseOperationCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_baseOperationCache_MultiGet(t *testing.T) {
	c := newBaseOperationCache()
	defer c.Close()

	var testData []*model.BaseOperation
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.BaseOperation))
	}

	err := c.ICache.(BaseOperationCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(BaseOperationCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.BaseOperation))
	}
}

func Test_baseOperationCache_MultiSet(t *testing.T) {
	c := newBaseOperationCache()
	defer c.Close()

	var testData []*model.BaseOperation
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.BaseOperation))
	}

	err := c.ICache.(BaseOperationCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_baseOperationCache_Del(t *testing.T) {
	c := newBaseOperationCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.BaseOperation)
	err := c.ICache.(BaseOperationCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_baseOperationCache_SetCacheWithNotFound(t *testing.T) {
	c := newBaseOperationCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.BaseOperation)
	err := c.ICache.(BaseOperationCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(BaseOperationCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewBaseOperationCache(t *testing.T) {
	c := NewBaseOperationCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewBaseOperationCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewBaseOperationCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
