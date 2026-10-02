package database

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/logger"
	"github.com/go-dev-frame/sponge/pkg/sgorm"
	"github.com/go-dev-frame/sponge/pkg/sgorm/postgresql"
	"github.com/go-dev-frame/sponge/pkg/utils"

	"semi-mes/server/internal/config"
)

// InitPostgresql connect postgresql. Switch the driver in configs/mes.yml to use it.
func InitPostgresql() *sgorm.DB {
	postgresqlCfg := config.Get().Database.Postgresql
	opts := []postgresql.Option{
		postgresql.WithMaxIdleConns(postgresqlCfg.MaxIdleConns),
		postgresql.WithMaxOpenConns(postgresqlCfg.MaxOpenConns),
		postgresql.WithConnMaxLifetime(time.Duration(postgresqlCfg.ConnMaxLifetime) * time.Minute),
	}
	if postgresqlCfg.EnableLog {
		opts = append(opts,
			postgresql.WithLogging(logger.Get()),
			postgresql.WithLogRequestIDKey("request_id"),
		)
	}

	if config.Get().App.EnableTrace {
		opts = append(opts, postgresql.WithEnableTrace())
	}

	dsn := utils.AdaptivePostgresqlDsn(postgresqlCfg.Dsn)
	db, err := postgresql.Init(dsn, opts...)
	if err != nil {
		panic("init postgresql error: " + err.Error())
	}
	return db
}
