package database

import (
	"os"
	"path/filepath"
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
	dbFile := utils.AdaptiveSqlite(sqliteCfg.DBFile)
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
