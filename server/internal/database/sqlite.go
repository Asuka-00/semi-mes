package database

import (
	"os"
	"path/filepath"
	"time"

	"github.com/go-dev-frame/sponge/pkg/logger"
	"github.com/go-dev-frame/sponge/pkg/sgorm"
	"github.com/go-dev-frame/sponge/pkg/sgorm/sqlite"
	"github.com/go-dev-frame/sponge/pkg/utils"

	"semi-mes/server/internal/config"
)

// InitSqlite connect sqlite
func InitSqlite() *sgorm.DB {
	sqliteCfg := config.Get().Database.Sqlite
	opts := []sqlite.Option{
		sqlite.WithMaxIdleConns(sqliteCfg.MaxIdleConns),
		sqlite.WithMaxOpenConns(sqliteCfg.MaxOpenConns),
		sqlite.WithConnMaxLifetime(time.Duration(sqliteCfg.ConnMaxLifetime) * time.Minute),
	}
	if sqliteCfg.EnableLog {
		opts = append(opts,
			sqlite.WithLogging(logger.Get()),
			sqlite.WithLogRequestIDKey("request_id"),
		)
	}

	if config.Get().App.EnableTrace {
		opts = append(opts, sqlite.WithEnableTrace())
	}

	dbFile := utils.AdaptiveSqlite(sqliteCfg.DBFile)
	if dir := filepath.Dir(dbFile); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			panic("create sqlite directory error: " + err.Error())
		}
	}
	db, err := sqlite.Init(dbFile, opts...)
	if err != nil {
		panic("init sqlite error: " + err.Error())
	}
	return db
}
