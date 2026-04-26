// Package logger provides a GORM logger adapter that routes all database logs
// through zerolog for consistent, structured output.
package logger

import (
	"context"
	"errors"
	"time"

	"github.com/rs/zerolog"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// GormLogger adapts zerolog for use as the GORM logger, ensuring database
// log output matches the application's standard structured format.
type GormLogger struct {
	zlog  zerolog.Logger
	level gormlogger.LogLevel
}

// NewGormLogger creates a GORM logger that writes through zerolog.
func NewGormLogger(zlog zerolog.Logger) *GormLogger {
	return &GormLogger{
		zlog:  zlog.With().Str("component", "gorm").Logger(),
		level: gormlogger.Warn,
	}
}

func (l *GormLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	return &GormLogger{zlog: l.zlog, level: level}
}

func (l *GormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.level >= gormlogger.Info {
		l.zlog.Info().Msgf(msg, data...)
	}
}

func (l *GormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.level >= gormlogger.Warn {
		l.zlog.Warn().Msgf(msg, data...)
	}
}

func (l *GormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.level >= gormlogger.Error {
		l.zlog.Error().Msgf(msg, data...)
	}
}

// Trace logs SQL queries with timing information.
// ErrRecordNotFound is silenced — it is a normal query result, not an error.
func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if l.level <= gormlogger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	switch {
	case err != nil && !errors.Is(err, gorm.ErrRecordNotFound) && l.level >= gormlogger.Error:
		l.zlog.Error().
			Err(err).
			Dur("elapsed", elapsed).
			Int64("rows", rows).
			Str("sql", sql).
			Msg("query error")

	case elapsed > 200*time.Millisecond && l.level >= gormlogger.Warn:
		l.zlog.Warn().
			Dur("elapsed", elapsed).
			Int64("rows", rows).
			Str("sql", sql).
			Msg("slow query")

	case l.level >= gormlogger.Info:
		l.zlog.Debug().
			Dur("elapsed", elapsed).
			Int64("rows", rows).
			Str("sql", sql).
			Msg("query")
	}
}
