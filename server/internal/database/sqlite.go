package database

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/go-dev-frame/sponge/pkg/sgorm"
	"github.com/go-dev-frame/sponge/pkg/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"

	"semi-mes/server/internal/config"
)

// InitSqlite opens SQLite with the pure-Go glebarez driver so CGO_ENABLED=0 builds can link.
func InitSqlite() *sgorm.DB {
	sqliteCfg := config.Get().Database.Sqlite
	dbFile := sqliteDSN(utils.AdaptiveSqlite(sqliteCfg.DBFile))
	if dir := filepath.Dir(dbFile); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			panic("create sqlite directory error: " + err.Error())
		}
	}

	logMode := logger.Silent
	if sqliteCfg.EnableLog {
		logMode = logger.Info
	}
	db, err := gorm.Open(sqlite.Open(dbFile), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		NamingStrategy:                           schema.NamingStrategy{SingularTable: true},
		Logger:                                   logger.Default.LogMode(logMode),
	})
	if err != nil {
		panic("init sqlite error: " + err.Error())
	}
	sqlDB, err := db.DB()
	if err != nil {
		panic("init sqlite error: " + err.Error())
	}
	if sqliteCfg.MaxIdleConns > 0 {
		sqlDB.SetMaxIdleConns(sqliteCfg.MaxIdleConns)
	}
	if sqliteCfg.MaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(sqliteCfg.MaxOpenConns)
	}
	if sqliteCfg.ConnMaxLifetime > 0 {
		sqlDB.SetConnMaxLifetime(time.Duration(sqliteCfg.ConnMaxLifetime) * time.Minute)
	}
	return db
}

// sqliteDSN asks the driver to begin write transactions immediately and to wait
// on a busy database. Deferred begins deadlock under concurrent sequence allocation.
func sqliteDSN(file string) string {
	file = strings.TrimSpace(file)
	if file == "" {
		file = "mes.db"
	}
	const params = "_txlock=immediate&_pragma=busy_timeout(5000)"
	if strings.Contains(file, "?") {
		return file + "&" + params
	}
	return file + "?" + params
}
