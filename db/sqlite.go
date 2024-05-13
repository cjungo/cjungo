package db

import (
	"os"
	"path/filepath"

	"github.com/rs/zerolog"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type SqliteConf struct {
	Path string
}

type Sqlite struct {
	*gorm.DB
}

type SqliteProvide func(*SqliteConf, *zerolog.Logger) (*Sqlite, error)

func NewSqliteHandle(initialize func(*Sqlite) error) SqliteProvide {
	return func(conf *SqliteConf, logger *zerolog.Logger) (*Sqlite, error) {
		db, err := gorm.Open(sqlite.Open(conf.Path), &gorm.Config{
			DisableForeignKeyConstraintWhenMigrating: true, // 禁止外键生成
			Logger: &DbLogger{
				sign:    "[SQLITE]",
				subject: logger,
			},
		})

		if err != nil {
			return nil, err
		}

		result := &Sqlite{DB: db}

		return result, initialize(result)
	}
}

func LoadSqliteConfFormEnv(logger *zerolog.Logger) (*SqliteConf, error) {
	conf := &SqliteConf{}
	logger.Info().Msg("SQLITE 通过环境变量加载配置")
	path := os.Getenv("CJUNGO_SQLITE_PATH")
	if len(path) == 0 {
		wd, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		conf.Path = filepath.Join(wd, "cjungo.db")
		logger.Info().Msg("SQLITE 使用默认配置")
	} else {
		conf.Path = path
	}

	logger.Info().Str("path", conf.Path).Msg("SQLITE 配置")

	return conf, nil
}
