package cjungo

import (
	"io"
	"os"
	"strings"

	"github.com/rs/zerolog"
	"go.uber.org/dig"
	"gopkg.in/natefinch/lumberjack.v2"
)

type LoggerConf struct {
	IsOutputConsole bool

	Filename   string
	MaxSize    *int
	MaxBackups *int
	MaxAge     *int
	IsCompress *bool
}

type NewLoggerDi struct {
	dig.In
	Conf *LoggerConf `optional:"true"`
}

func NewLogger(di NewLoggerDi) *zerolog.Logger {
	writers := []io.Writer{}

	// 提供配置
	if di.Conf != nil {
		if di.Conf.IsOutputConsole {
			consoleWriter := zerolog.ConsoleWriter{
				Out: os.Stdout,
			}
			writers = append(writers, consoleWriter)
		}

		if strings.HasSuffix(di.Conf.Filename, ".log") {
			lumberLogger := &lumberjack.Logger{
				Filename:   di.Conf.Filename,
				MaxSize:    GetOrDefault(di.Conf.MaxSize, 4),
				MaxBackups: GetOrDefault(di.Conf.MaxBackups, 3),
				MaxAge:     GetOrDefault(di.Conf.MaxAge, 14),
				Compress:   GetOrDefault(di.Conf.IsCompress, true),
			}
			writers = append(writers, lumberLogger)
		}
	} else { // 默认配置
		consoleWriter := zerolog.ConsoleWriter{
			Out: os.Stdout,
		}
		writers = append(writers, consoleWriter)
	}

	multiWriter := zerolog.MultiLevelWriter(writers...)
	logger := zerolog.New(multiWriter).With().Timestamp().Logger()
	return &logger
}

func LoadLoggerConfFromEnv() (*LoggerConf, error) {
	conf := &LoggerConf{
		IsOutputConsole: true,
	}
	filename := os.Getenv("CJUNGO_LOG_FILENAME")
	if strings.HasSuffix(filename, ".log") {
		conf.Filename = filename
	}

	if err := GetEnvBool("CJUNGO_LOG_IS_OUTPUT_CONSOLE", func(v bool) {
		conf.IsOutputConsole = v
	}); err != nil {
		return nil, err
	}

	if err := GetEnvInt("CJUNGO_LOG_MAX_SIZE", func(v int) {
		conf.MaxSize = &v
	}); err != nil {
		return nil, err
	}
	if err := GetEnvInt("CJUNGO_LOG_MAX_BACKUPS", func(v int) {
		conf.MaxBackups = &v
	}); err != nil {
		return nil, err
	}
	if err := GetEnvInt("CJUNGO_LOG_MAX_AGE", func(v int) {
		conf.MaxAge = &v
	}); err != nil {
		return nil, err
	}
	if err := GetEnvBool("CJUNGO_LOG_IS_COMPRESS", func(v bool) {
		conf.IsCompress = &v
	}); err != nil {
		return nil, err
	}

	return conf, nil
}
