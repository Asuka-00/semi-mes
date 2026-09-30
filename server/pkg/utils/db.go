package utils

import (
	"fmt"
	"os"
	"path/filepath"

	"semi-mes/server/internal/config"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func InitDB(cfg *config.Config) (*gorm.DB, error) {
	var dialector gorm.Dialector

	switch cfg.Database.Type {
	case "sqlite":
		dbPath := cfg.Database.SQLite.Path
		dir := filepath.Dir(dbPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create data directory: %w", err)
		}
		dialector = sqlite.Open(dbPath)
	case "mysql":
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
			cfg.Database.MySQL.Username,
			cfg.Database.MySQL.Password,
			cfg.Database.MySQL.Host,
			cfg.Database.MySQL.Port,
			cfg.Database.MySQL.Database,
			cfg.Database.MySQL.Charset,
		)
		dialector = mysql.Open(dsn)
	case "postgres":
		dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			cfg.Database.Postgres.Host,
			cfg.Database.Postgres.Port,
			cfg.Database.Postgres.Username,
			cfg.Database.Postgres.Password,
			cfg.Database.Postgres.Database,
			cfg.Database.Postgres.SSLMode,
		)
		dialector = postgres.Open(dsn)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", cfg.Database.Type)
	}

	gormConfig := &gorm.Config{}
	if cfg.Log.Level == "debug" {
		gormConfig.Logger = logger.Default.LogMode(logger.Info)
	} else {
		gormConfig.Logger = logger.Default.LogMode(logger.Silent)
	}

	db, err := gorm.Open(dialector, gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return db, nil
}
