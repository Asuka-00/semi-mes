package dao

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-dev-frame/sponge/pkg/gotest"
	"github.com/go-dev-frame/sponge/pkg/sgorm/query"
	"github.com/go-dev-frame/sponge/pkg/utils"
	"github.com/stretchr/testify/assert"

	"semi-mes/server/internal/cache"
	"semi-mes/server/internal/database"
	"semi-mes/server/internal/model"
)

func newBaseProductionLineDao() *gotest.Dao {
	testData := &model.BaseProductionLine{}
	testData.ID = 1
	// you can set the other fields of testData here, such as:
	//testData.CreatedAt = time.Now()
	//testData.UpdatedAt = testData.CreatedAt

	// init mock cache
	//c := gotest.NewCache(map[string]interface{}{"no cache": testData}) // to test mysql, disable caching
	c := gotest.NewCache(map[string]interface{}{utils.Uint64ToStr(testData.ID): testData})
	c.ICache = cache.NewBaseProductionLineCache(&database.CacheType{
		CType: "redis",
		Rdb:   c.RedisClient,
	})

	// init mock dao
	d := gotest.NewDao(c, testData)
	d.IDao = NewBaseProductionLineDao(d.DB, c.ICache.(cache.BaseProductionLineCache))

	return d
}

func Test_baseProductionLineDao_Create(t *testing.T) {
	d := newBaseProductionLineDao()
	defer d.Close()
	testData := d.TestData.(*model.BaseProductionLine)

	d.SQLMock.ExpectBegin()
	insertArgs := d.GetAnyArgs(testData)
	insertArgs = insertArgs[:len(insertArgs)-1] // gorm.DeletedAt is one column but reflects as two fields
	d.SQLMock.ExpectExec("INSERT INTO .*").
		WithArgs(insertArgs...).
		WillReturnResult(sqlmock.NewResult(1, 1))
	d.SQLMock.ExpectCommit()

	err := d.IDao.(BaseProductionLineDao).Create(d.Ctx, testData)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_baseProductionLineDao_DeleteByID(t *testing.T) {
	d := newBaseProductionLineDao()
	defer d.Close()
	testData := d.TestData.(*model.BaseProductionLine)
	expectedSQLForDeletion := "UPDATE .*"
	

	d.SQLMock.ExpectBegin()
	d.SQLMock.ExpectExec(expectedSQLForDeletion).
		WithArgs(d.AnyTime, testData.ID).
		WillReturnResult(sqlmock.NewResult(int64(testData.ID), 1))
	d.SQLMock.ExpectCommit()

	err := d.IDao.(BaseProductionLineDao).DeleteByID(d.Ctx, testData.ID)
	if err != nil {
		t.Fatal(err)
	}

	// zero id error
	err = d.IDao.(BaseProductionLineDao).DeleteByID(d.Ctx, 0)
	assert.Error(t, err)
}

func Test_baseProductionLineDao_UpdateByID(t *testing.T) {
	d := newBaseProductionLineDao()
	defer d.Close()
	testData := d.TestData.(*model.BaseProductionLine)

	d.SQLMock.ExpectBegin()
	d.SQLMock.ExpectExec("UPDATE .*").
		WillReturnResult(sqlmock.NewResult(1, 1))
	d.SQLMock.ExpectCommit()

	err := d.IDao.(BaseProductionLineDao).UpdateByID(d.Ctx, testData)
	if err != nil {
		t.Fatal(err)
	}

	// zero id error
	err = d.IDao.(BaseProductionLineDao).UpdateByID(d.Ctx, &model.BaseProductionLine{})
	assert.Error(t, err)
	
}

func Test_baseProductionLineDao_GetByID(t *testing.T) {
	d := newBaseProductionLineDao()
	defer d.Close()
	testData := d.TestData.(*model.BaseProductionLine)

	// column names and corresponding data
	rows := sqlmock.NewRows([]string{"id"}).
		AddRow(testData.ID)

	d.SQLMock.ExpectQuery("SELECT .*").
		WithArgs(testData.ID, 1).
		WillReturnRows(rows)

	_, err := d.IDao.(BaseProductionLineDao).GetByID(d.Ctx, testData.ID)
	if err != nil {
		t.Fatal(err)
	}

	err = d.SQLMock.ExpectationsWereMet()
	if err != nil {
		t.Fatal(err)
	}

	// error test
	d.SQLMock.ExpectQuery("SELECT .*").
		WithArgs(2).
		WillReturnRows(rows)
	_, err = d.IDao.(BaseProductionLineDao).GetByID(d.Ctx, 2)
	assert.Error(t, err)

	d.SQLMock.ExpectQuery("SELECT .*").
		WithArgs(3, 4).
		WillReturnRows(rows)
	_, err = d.IDao.(BaseProductionLineDao).GetByID(d.Ctx, 4)
	assert.Error(t, err)
}

func Test_baseProductionLineDao_GetByColumns(t *testing.T) {
	d := newBaseProductionLineDao()
	defer d.Close()
	testData := d.TestData.(*model.BaseProductionLine)

	// column names and corresponding data
	rows := sqlmock.NewRows([]string{"id"}).
		AddRow(testData.ID)

	d.SQLMock.ExpectQuery("SELECT .*").WillReturnRows(rows)

	_, _, err := d.IDao.(BaseProductionLineDao).GetByColumns(d.Ctx, &query.Params{
		Page:  0,
		Limit: 10,
		Sort:  "ignore count", // ignore test count(*)
	})
	if err != nil {
		t.Fatal(err)
	}

	err = d.SQLMock.ExpectationsWereMet()
	if err != nil {
		t.Fatal(err)
	}

	// err test
	_, _, err = d.IDao.(BaseProductionLineDao).GetByColumns(d.Ctx, &query.Params{
		Page:  0,
		Limit: 10,
		Columns: []query.Column{
			{
				Name:  "id",
				Exp:   "<",
				Value: 0,
			},
		},
	})
	assert.Error(t, err)

	// error test
	dao := &baseProductionLineDao{}
	_, _, err = dao.GetByColumns(context.Background(), &query.Params{Columns: []query.Column{{}}})
	t.Log(err)
}

func Test_baseProductionLineDao_CreateByTx(t *testing.T) {
	d := newBaseProductionLineDao()
	defer d.Close()
	testData := d.TestData.(*model.BaseProductionLine)

	d.SQLMock.ExpectBegin()
	insertArgs := d.GetAnyArgs(testData)
	insertArgs = insertArgs[:len(insertArgs)-1] // gorm.DeletedAt is one column but reflects as two fields
	d.SQLMock.ExpectExec("INSERT INTO .*").
		WithArgs(insertArgs...).
		WillReturnResult(sqlmock.NewResult(1, 1))
	d.SQLMock.ExpectCommit()

	_, err := d.IDao.(BaseProductionLineDao).CreateByTx(d.Ctx, d.DB, testData)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_baseProductionLineDao_DeleteByTx(t *testing.T) {
	d := newBaseProductionLineDao()
	defer d.Close()
	testData := d.TestData.(*model.BaseProductionLine)
	expectedSQLForDeletion := "UPDATE .*"
	

	d.SQLMock.ExpectBegin()
	d.SQLMock.ExpectExec(expectedSQLForDeletion).
		WithArgs(d.AnyTime, testData.ID).
		WillReturnResult(sqlmock.NewResult(int64(testData.ID), 1))
	d.SQLMock.ExpectCommit()

	err := d.IDao.(BaseProductionLineDao).DeleteByTx(d.Ctx, d.DB, testData.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_baseProductionLineDao_UpdateByTx(t *testing.T) {
	d := newBaseProductionLineDao()
	defer d.Close()
	testData := d.TestData.(*model.BaseProductionLine)

	d.SQLMock.ExpectBegin()
	d.SQLMock.ExpectExec("UPDATE .*").
		WillReturnResult(sqlmock.NewResult(1, 1))
	d.SQLMock.ExpectCommit()

	err := d.IDao.(BaseProductionLineDao).UpdateByTx(d.Ctx, d.DB, testData)
	if err != nil {
		t.Fatal(err)
	}
}
