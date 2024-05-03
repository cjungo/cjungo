package cjungo

import (
	"io"
	"os"
	"strings"

	"github.com/rs/zerolog"
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

func NewLogger(conf *LoggerConf) *zerolog.Logger {
	writers := []io.Writer{}

	if conf.IsOutputConsole {
		consoleWriter := zerolog.ConsoleWriter{
			Out: os.Stdout,
		}
		writers = append(writers, consoleWriter)
	}

	if strings.HasSuffix(conf.Filename, ".log") {
		lumberLogger := &lumberjack.Logger{
			Filename:   conf.Filename,
			MaxSize:    GetOrDefault(conf.MaxSize, 4),
			MaxBackups: GetOrDefault(conf.MaxBackups, 3),
			MaxAge:     GetOrDefault(conf.MaxAge, 14),
			Compress:   GetOrDefault(conf.IsCompress, true),
		}
		writers = append(writers, lumberLogger)
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
