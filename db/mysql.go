package db

import (
	"context"
	"fmt"
	"time"

	"github.com/cjungo/cjungo"
	"github.com/rs/zerolog"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	glog "gorm.io/gorm/logger"
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

type MySqlLogger struct {
	subject *zerolog.Logger
}

func (logger *MySqlLogger) LogMode(level glog.LogLevel) glog.Interface {
	logger.subject.Info().Msg("[MYSQL] LogMode")
	return logger
}
func (logger *MySqlLogger) Info(ctx context.Context, f string, a ...any) {
	logger.subject.Info().Msg("[MYSQL] Info")
	logger.subject.Info().Msgf(f, a...)
}
func (logger *MySqlLogger) Warn(ctx context.Context, f string, a ...any) {
	logger.subject.Info().Msg("[MYSQL] Warn")
	logger.subject.Warn().Msgf(f, a...)
}
func (logger *MySqlLogger) Error(ctx context.Context, f string, a ...any) {
	logger.subject.Info().Msg("[MYSQL] Error")
	logger.subject.Error().Msgf(f, a...)
}
func (logger *MySqlLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	sql, rowsAffected := fc()
	logger.subject.Info().
		Time("begin", begin).
		Err(err).
		Str("sql", sql).
		Int64("rowsAffected", rowsAffected).
		Msg("[MYSQL]")
}

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
			Logger: &MySqlLogger{subject: logger},
		})
		if err != nil {
			return nil, err
		}

		result := &MySql{DB: db}
		return result, initialize(result)
	}
}

func LoadMySqlConfFormEnv(logger *zerolog.Logger) (*MySqlConf, error) {
	conf := &MySqlConf{}

	logger.Info().Msg("MYSQL 通过环境变量加载配置")

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
		logger.Info().Msg("MySql 使用默认端口")
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
