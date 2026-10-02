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

func newBaseRecipeCache() *gotest.Cache {
	record1 := &model.BaseRecipe{}
	record1.ID = 1
	record2 := &model.BaseRecipe{}
	record2.ID = 2
	testData := map[string]interface{}{
		utils.Uint64ToStr(record1.ID): record1,
		utils.Uint64ToStr(record2.ID): record2,
	}

	c := gotest.NewCache(testData)
	c.ICache = NewBaseRecipeCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})
	return c
}

func Test_baseRecipeCache_Set(t *testing.T) {
	c := newBaseRecipeCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.BaseRecipe)
	err := c.ICache.(BaseRecipeCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// nil data
	err = c.ICache.(BaseRecipeCache).Set(c.Ctx, 0, nil, time.Hour)
	assert.NoError(t, err)
}

func Test_baseRecipeCache_Get(t *testing.T) {
	c := newBaseRecipeCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.BaseRecipe)
	err := c.ICache.(BaseRecipeCache).Set(c.Ctx, record.ID, record, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(BaseRecipeCache).Get(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, record, got)

	// zero key error
	_, err = c.ICache.(BaseRecipeCache).Get(c.Ctx, 0)
	assert.Error(t, err)
}

func Test_baseRecipeCache_MultiGet(t *testing.T) {
	c := newBaseRecipeCache()
	defer c.Close()

	var testData []*model.BaseRecipe
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.BaseRecipe))
	}

	err := c.ICache.(BaseRecipeCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.ICache.(BaseRecipeCache).MultiGet(c.Ctx, c.GetIDs())
	if err != nil {
		t.Fatal(err)
	}

	expected := c.GetTestData()
	for k, v := range expected {
		assert.Equal(t, got[utils.StrToUint64(k)], v.(*model.BaseRecipe))
	}
}

func Test_baseRecipeCache_MultiSet(t *testing.T) {
	c := newBaseRecipeCache()
	defer c.Close()

	var testData []*model.BaseRecipe
	for _, data := range c.TestDataSlice {
		testData = append(testData, data.(*model.BaseRecipe))
	}

	err := c.ICache.(BaseRecipeCache).MultiSet(c.Ctx, testData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_baseRecipeCache_Del(t *testing.T) {
	c := newBaseRecipeCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.BaseRecipe)
	err := c.ICache.(BaseRecipeCache).Del(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_baseRecipeCache_SetCacheWithNotFound(t *testing.T) {
	c := newBaseRecipeCache()
	defer c.Close()

	record := c.TestDataSlice[0].(*model.BaseRecipe)
	err := c.ICache.(BaseRecipeCache).SetPlaceholder(c.Ctx, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	b := c.ICache.(BaseRecipeCache).IsPlaceholderErr(err)
	t.Log(b)
}

func TestNewBaseRecipeCache(t *testing.T) {
	c := NewBaseRecipeCache(&database.CacheType{
		CType: "",
	})
	assert.Nil(t, c)
	c = NewBaseRecipeCache(&database.CacheType{
		CType: "memory",
	})
	assert.NotNil(t, c)
	c = NewBaseRecipeCache(&database.CacheType{
		CType: "redis",
	})
	assert.NotNil(t, c)
}
