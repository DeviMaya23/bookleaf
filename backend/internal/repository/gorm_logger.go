package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/devi/bookleaf/internal/observability"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const defaultSlowQueryThreshold = 200 * time.Millisecond

type zapGORMLogger struct {
	base      *zap.Logger
	level     logger.LogLevel
	threshold time.Duration
}

func NewZapGORMLogger(base *zap.Logger) logger.Interface {
	return &zapGORMLogger{
		base:      base,
		level:     logger.Warn,
		threshold: defaultSlowQueryThreshold,
	}
}

func (l *zapGORMLogger) LogMode(level logger.LogLevel) logger.Interface {
	clone := *l
	clone.level = level
	return &clone
}

func (l *zapGORMLogger) Info(_ context.Context, msg string, args ...interface{}) {
	if l.level >= logger.Info {
		l.base.Debug(fmt.Sprintf(msg, args...))
	}
}

func (l *zapGORMLogger) Warn(_ context.Context, msg string, args ...interface{}) {
	if l.level >= logger.Warn {
		l.base.Warn(fmt.Sprintf(msg, args...))
	}
}

func (l *zapGORMLogger) Error(_ context.Context, msg string, args ...interface{}) {
	if l.level >= logger.Error {
		l.base.Error(fmt.Sprintf(msg, args...))
	}
}

func (l *zapGORMLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.level <= logger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rowsAffected := fc()
	log := observability.LoggerFromContext(ctx, l.base)

	switch {
	case err != nil && err != gorm.ErrRecordNotFound:
		log.Error("gorm error",
			zap.Error(err),
			zap.Float64("elapsed_ms", float64(elapsed.Milliseconds())),
			zap.Int64("rows_affected", rowsAffected),
			zap.String("sql", sql),
		)
	case elapsed > l.threshold:
		log.Warn("slow query",
			zap.Float64("elapsed_ms", float64(elapsed.Milliseconds())),
			zap.Int64("rows_affected", rowsAffected),
			zap.String("sql", sql),
		)
	}
}
