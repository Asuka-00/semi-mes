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

func newBaseProductionLineCache() *gotest.Cache {
	record1 := &model.BaseProductionLine{}
	record1.ID = 1
	record2 := &model.BaseProductionLine{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewBaseProductionLineCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_baseProductionLineCache_Set(t *testing.T) {
	c := newBaseProductionLineCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.BaseProductionLine)
	err := c.ICache.(BaseProductionLineCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(BaseProductionLineCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_baseProductionLineCache_Get(t *testing.T) {
	c := newBaseProductionLineCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.BaseProductionLine)
	err := c.ICache.(BaseProductionLineCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(BaseProductionLineCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(BaseProductionLineCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_baseProductionLineCache_MultiGet(t *testing.T) {
	c := newBaseProductionLineCache()
	defer c.Close()

	var testData []*model.BaseProductionLine
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.BaseProductionLine))
	}

	err := c.ICache.(BaseProductionLineCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(BaseProductionLineCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.BaseProductionLine))
	}
}

func Test_baseProductionLineCache_MultiSet(t *testing.T) {
	c := newBaseProductionLineCache()
	defer c.Close()

	var testData []*model.BaseProductionLine
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.BaseProductionLine))
	}

	err := c.ICache.(BaseProductionLineCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_baseProductionLineCache_Del(t *testing.T) {
	c := newBaseProductionLineCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.BaseProductionLine)
	err := c.ICache.(BaseProductionLineCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_baseProductionLineCache_SetCacheWithNotFound(t *testing.T) {
	c := newBaseProductionLineCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.BaseProductionLine)
	err := c.ICache.(BaseProductionLineCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(BaseProductionLineCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewBaseProductionLineCache(t *testing.T) {
	c := NewBaseProductionLineCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewBaseProductionLineCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewBaseProductionLineCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
