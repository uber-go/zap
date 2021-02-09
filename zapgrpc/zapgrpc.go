// Copyright (c) 2016 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

// Package zapgrpc provides a logger that is compatible with grpclog.
package zapgrpc // import "go.uber.org/zap/zapgrpc"

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// See https://github.com/grpc/grpc-go/blob/v1.35.0/grpclog/loggerv2.go#L77-L86
const (
	grpcLvlInfo  = 0
	grpcLvlWarn  = 1
	grpcLvlError = 2
	grpcLvlFatal = 3
)

var (
	// _grpcToZapLevel maps gRPC log levels to zap log levels.
	// See https://pkg.go.dev/go.uber.org/zap@v1.16.0/zapcore#Level
	_grpcToZapLevel = map[int]zapcore.Level{
		grpcLvlInfo:  zapcore.InfoLevel,
		grpcLvlWarn:  zapcore.WarnLevel,
		grpcLvlError: zapcore.ErrorLevel,
		grpcLvlFatal: zapcore.FatalLevel,
	}
)

// An Option overrides a Logger's default configuration.
type Option interface {
	apply(*Logger)
}

type optionFunc func(*Logger)

func (f optionFunc) apply(log *Logger) {
	f(log)
}

// WithDebug configures a Logger to print at zap's DebugLevel instead of
// InfoLevel.
// It only affects the Printf, Println and Print methods, which are only used in the gRPC v1 grpclog.Logger API.
// Deprecated: use grpclog.SetLoggerV2() for v2 API.
func WithDebug() Option {
	return optionFunc(func(logger *Logger) {
		logger.printToDebug = true
	})
}

// NewLogger returns a new Logger.
func NewLogger(l *zap.Logger, options ...Option) *Logger {
	logger := &Logger{
		delegate:     l.Sugar(),
		levelEnabler: l.Core(),
		printToDebug: false,
		fatalToWarn:  false,
	}
	for _, option := range options {
		option.apply(logger)
	}
	return logger
}

// Logger adapts zap's Logger to be compatible with grpclog.LoggerV2 and the deprecated grpclog.Logger.
type Logger struct {
	delegate     *zap.SugaredLogger
	levelEnabler zapcore.LevelEnabler
	printToDebug bool
	fatalToWarn  bool
}

// Print implements grpclog.Logger.
// Deprecated: use Info().
func (l *Logger) Print(args ...interface{}) {
	if l.printToDebug {
		l.delegate.Debug(args...)
	} else {
		l.delegate.Info(args...)
	}
}

// Printf implements grpclog.Logger.
// Deprecated: use Infof().
func (l *Logger) Printf(format string, args ...interface{}) {
	if l.printToDebug {
		l.delegate.Debugf(format, args...)
	} else {
		l.delegate.Infof(format, args...)
	}
}

// Println implements grpclog.Logger.
// Deprecated: use Info().
func (l *Logger) Println(args ...interface{}) {
	if l.printToDebug {
		if l.levelEnabler.Enabled(zapcore.DebugLevel) {
			l.delegate.Debug(sprintln(args))
		}
	} else {
		if l.levelEnabler.Enabled(zapcore.InfoLevel) {
			l.delegate.Info(sprintln(args))
		}
	}
}

// Info implements grpclog.LoggerV2.
func (l *Logger) Info(args ...interface{}) {
	l.delegate.Info(args...)
}

// Infoln implements grpclog.LoggerV2.
func (l *Logger) Infoln(args ...interface{}) {
	if l.levelEnabler.Enabled(zapcore.InfoLevel) {
		l.delegate.Info(sprintln(args))
	}
}

// Infof implements grpclog.LoggerV2.
func (l *Logger) Infof(format string, args ...interface{}) {
	l.delegate.Infof(format, args...)
}

// Warning implements grpclog.LoggerV2.
func (l *Logger) Warning(args ...interface{}) {
	l.delegate.Warn(args...)
}

// Warningln implements grpclog.LoggerV2.
func (l *Logger) Warningln(args ...interface{}) {
	if l.levelEnabler.Enabled(zapcore.WarnLevel) {
		l.delegate.Warn(sprintln(args))
	}
}

// Warningf implements grpclog.LoggerV2.
func (l *Logger) Warningf(format string, args ...interface{}) {
	l.delegate.Warnf(format, args...)
}

// Error implements grpclog.LoggerV2.
func (l *Logger) Error(args ...interface{}) {
	l.delegate.Error(args...)
}

// Errorln implements grpclog.LoggerV2.
func (l *Logger) Errorln(args ...interface{}) {
	if l.levelEnabler.Enabled(zapcore.ErrorLevel) {
		l.delegate.Error(sprintln(args))
	}
}

// Errorf implements grpclog.LoggerV2.
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.delegate.Errorf(format, args...)
}

// Fatal implements grpclog.LoggerV2.
func (l *Logger) Fatal(args ...interface{}) {
	if l.fatalToWarn {
		l.delegate.Warn(args...)
	} else {
		l.delegate.Fatal(args...)
	}
}

// Fatalln implements grpclog.LoggerV2.
func (l *Logger) Fatalln(args ...interface{}) {
	if l.fatalToWarn {
		if l.levelEnabler.Enabled(zapcore.WarnLevel) {
			l.delegate.Warn(sprintln(args))
		}
	} else {
		if l.levelEnabler.Enabled(zapcore.FatalLevel) {
			l.delegate.Fatal(sprintln(args))
		}
	}
}

// Fatalf implements grpclog.LoggerV2.
func (l *Logger) Fatalf(format string, args ...interface{}) {
	if l.fatalToWarn {
		l.delegate.Warnf(format, args...)
	} else {
		l.delegate.Fatalf(format, args...)
	}
}

// V implements grpclog.LoggerV2.
func (l *Logger) V(level int) bool {
	return l.levelEnabler.Enabled(_grpcToZapLevel[level])
}

func sprintln(args []interface{}) string {
	s := fmt.Sprintln(args...)
	return s[:len(s)-1]
}
