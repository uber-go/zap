// Package lazylogger provides an experimental logger that supports
// lazy with operations
package lazylogger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// LazyLogger is a proposed breaking change to logger that
// adds a WithLazy method
type LazyLogger struct {
	*zap.Logger
}

// WithLazy creates a child logger and lazily encodes structured context if the
// child logger is ever further chained with With() or is written to with any
// of the log level methods. Until the occurs, the logger may retain references
// to references in objects, etc.
// Fields added to the child don't affect the parent, and vice versa.
func (log *LazyLogger) WithLazy(fields ...zap.Field) *LazyLogger {
	if len(fields) == 0 {
		return log
	}
	return &LazyLogger{
		log.WithOptions(zap.WrapCore(func(core zapcore.Core) zapcore.Core {
			return zapcore.NewLazyWith(core, fields)
		})),
	}
}

// With is for testing in exp package only
func (log *LazyLogger) With(fields ...zap.Field) *LazyLogger {
	return &LazyLogger{
		log.Logger.With(fields...),
	}
}

// New is for testing in exp package only
func New(core zapcore.Core, options ...zap.Option) *LazyLogger {
	return &LazyLogger{
		zap.New(core, options...),
	}
}
