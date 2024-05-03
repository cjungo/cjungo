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
			MaxSize:    GetOrDefault(conf.MaxSize, 500),
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

func LoadLoggerConfFromEnv() *LoggerConf {
	return &LoggerConf{
		IsOutputConsole: true,
		Filename:        "./log/test.log",
	}
}
