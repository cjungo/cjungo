package db

import (
	"fmt"
	"time"

	"github.com/cjungo/cjungo"
	"github.com/rs/zerolog"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
)

type MySqlConf struct {
	Host string
	Port uint16
	User string
	Pass string
	Name string
}

type MySql struct {
	*gorm.DB
}

type MySqlProvide func(*MySqlConf, *zerolog.Logger) (*MySql, error)

func NewMySqlHandle(initialize func(*MySql) error) MySqlProvide {
	return func(conf *MySqlConf, logger *zerolog.Logger) (*MySql, error) {
		dns := fmt.Sprintf(
			"%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			conf.User,
			conf.Pass,
			conf.Host,
			conf.Port,
			conf.Name,
		)
		db, err := gorm.Open(mysql.Open(dns), &gorm.Config{
			DisableForeignKeyConstraintWhenMigrating: true, // 禁止外键生成
			Logger: &DbLogger{
				sign:    "[MYSQL]",
				subject: logger,
			},
		})
		if err != nil {
			return nil, err
		}
		if err = db.Use(
			dbresolver.Register(dbresolver.Config{}).
				SetConnMaxIdleTime(30 * time.Minute).
				SetConnMaxLifetime(2 * time.Hour).
				SetMaxIdleConns(32).
				SetMaxOpenConns(64),
		); err != nil {
			return nil, err
		}

		result := &MySql{DB: db}
		return result, initialize(result)
	}
}

func LoadMySqlConfFormEnv(logger *zerolog.Logger) (*MySqlConf, error) {
	conf := &MySqlConf{}

	logger.Info().Str("action", "通过环境变量加载配置").Msg("[MYSQL]")

	if err := cjungo.GetEnvStringMust("CJUNGO_MYSQL_HOST", func(v string) {
		conf.Host = v
	}); err != nil {
		return nil, err
	}

	if err := cjungo.GetEnvInt("CJUNGO_MYSQL_PORT", func(v uint16) {
		conf.Port = v
	}); err != nil {
		return nil, err
	} else {
		conf.Port = 3306
		logger.Info().Str("action", "使用默认端口").Msg("[MYSQL]")
	}

	if err := cjungo.GetEnvStringMust("CJUNGO_MYSQL_USER", func(v string) {
		conf.User = v
	}); err != nil {
		return nil, err
	}

	if err := cjungo.GetEnvStringMust("CJUNGO_MYSQL_PASS", func(v string) {
		conf.Pass = v
	}); err != nil {
		return nil, err
	}

	if err := cjungo.GetEnvStringMust("CJUNGO_MYSQL_NAME", func(v string) {
		conf.Name = v
	}); err != nil {
		return nil, err
	}

	return conf, nil
}
