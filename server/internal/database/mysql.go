package database

import (
	"time"

	"github.com/go-dev-frame/sponge/pkg/logger"
	"github.com/go-dev-frame/sponge/pkg/sgorm"
	"github.com/go-dev-frame/sponge/pkg/sgorm/mysql"
	"github.com/go-dev-frame/sponge/pkg/utils"

	"semi-mes/server/internal/config"
)

// InitMysql connect mysql. Switch the driver in configs/mes.yml to use it.
func InitMysql() *sgorm.DB {
	mysqlCfg := config.Get().Database.Mysql
	opts := []mysql.Option{
		mysql.WithMaxIdleConns(mysqlCfg.MaxIdleConns),
		mysql.WithMaxOpenConns(mysqlCfg.MaxOpenConns),
		mysql.WithConnMaxLifetime(time.Duration(mysqlCfg.ConnMaxLifetime) * time.Minute),
	}
	if mysqlCfg.EnableLog {
		opts = append(opts,
			mysql.WithLogging(logger.Get()),
			mysql.WithLogRequestIDKey("request_id"),
		)
	}

	if config.Get().App.EnableTrace {
		opts = append(opts, mysql.WithEnableTrace())
	}

	dsn := utils.AdaptiveMysqlDsn(mysqlCfg.Dsn)
	db, err := mysql.Init(dsn, opts...)
	if err != nil {
		panic("init mysql error: " + err.Error())
	}
	return db
}
