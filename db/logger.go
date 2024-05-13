package db

import (
	"context"
	"time"

	"github.com/rs/zerolog"
	glog "gorm.io/gorm/logger"
)

type DbLogger struct {
	sign    string
	subject *zerolog.Logger
}

func (logger *DbLogger) LogMode(level glog.LogLevel) glog.Interface {
	logger.subject.Info().Str("call", "LogMode").Msg(logger.sign)
	return logger
}
func (logger *DbLogger) Info(ctx context.Context, f string, a ...any) {
	logger.subject.Info().Str("call", "Info").Msg(logger.sign)
	logger.subject.Info().Msgf(f, a...)
}
func (logger *DbLogger) Warn(ctx context.Context, f string, a ...any) {
	logger.subject.Info().Str("call", "Warn").Msg(logger.sign)
	logger.subject.Warn().Msgf(f, a...)
}
func (logger *DbLogger) Error(ctx context.Context, f string, a ...any) {
	logger.subject.Info().Str("call", "Error").Msg(logger.sign)
	logger.subject.Error().Msgf(f, a...)
}
func (logger *DbLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	sql, rowsAffected := fc()
	logger.subject.Info().
		Time("begin", begin).
		Err(err).
		Str("sql", sql).
		Int64("rowsAffected", rowsAffected).
		Msg(logger.sign)
}
