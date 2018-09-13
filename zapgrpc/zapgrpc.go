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
	"go.uber.org/zap"
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
func WithDebug() Option {
	return optionFunc(func(logger *Logger) {
		logger.print = (*zap.SugaredLogger).Debug
		logger.printf = (*zap.SugaredLogger).Debugf
	})
}

// WithVerbosity sets verbose level referred by V().
func WithVerbosity(v int) Option {
	return optionFunc(func(logger *Logger) {
		logger.v = v
	})
}

// NewLogger returns a new Logger.
//
// By default, Loggers print at zap's InfoLevel.
func NewLogger(l *zap.Logger, options ...Option) *Logger {
	logger := &Logger{
		log:      l.Sugar(),
		info:     (*zap.SugaredLogger).Info,
		infof:    (*zap.SugaredLogger).Infof,
		warning:  (*zap.SugaredLogger).Warn,
		warningf: (*zap.SugaredLogger).Warnf,
		err:      (*zap.SugaredLogger).Error,
		errf:     (*zap.SugaredLogger).Errorf,
		fatal:    (*zap.SugaredLogger).Fatal,
		fatalf:   (*zap.SugaredLogger).Fatalf,
		print:    (*zap.SugaredLogger).Info,
		printf:   (*zap.SugaredLogger).Infof,
	}
	for _, option := range options {
		option.apply(logger)
	}
	return logger
}

// Logger adapts zap's Logger to be compatible with grpclog.Logger and grpclog.LoggerV2.
type Logger struct {
	v        int
	log      *zap.SugaredLogger
	info     func(*zap.SugaredLogger, ...interface{})
	infof    func(*zap.SugaredLogger, string, ...interface{})
	warning  func(*zap.SugaredLogger, ...interface{})
	warningf func(*zap.SugaredLogger, string, ...interface{})
	err      func(*zap.SugaredLogger, ...interface{})
	errf     func(*zap.SugaredLogger, string, ...interface{})
	fatal    func(*zap.SugaredLogger, ...interface{})
	fatalf   func(*zap.SugaredLogger, string, ...interface{})
	print    func(*zap.SugaredLogger, ...interface{})
	printf   func(*zap.SugaredLogger, string, ...interface{})
}

// Info implements grpclog.LoggerV2.
func (l *Logger) Info(args ...interface{}) {
	l.info(l.log, args...)
}

// Infof implements grpclog.LoggerV2.
func (l *Logger) Infof(format string, args ...interface{}) {
	l.infof(l.log, format, args...)
}

// Infoln implements grpclog.LoggerV2.
func (l *Logger) Infoln(args ...interface{}) {
	l.info(l.log, args...)
}

// Warning implements grpclog.LoggerV2.
func (l *Logger) Warning(args ...interface{}) {
	l.warning(l.log, args...)
}

// Warningf implements grpclog.LoggerV2.
func (l *Logger) Warningf(format string, args ...interface{}) {
	l.warningf(l.log, format, args...)
}

// Warningln implements grpclog.LoggerV2.
func (l *Logger) Warningln(args ...interface{}) {
	l.warning(l.log, args...)
}

// Error implements grpclog.LoggerV2.
func (l *Logger) Error(args ...interface{}) {
	l.err(l.log, args...)
}

// Errorf implements grpclog.LoggerV2.
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.errf(l.log, format, args...)
}

// Errorln implements grpclog.LoggerV2.
func (l *Logger) Errorln(args ...interface{}) {
	l.err(l.log, args...)
}

// Fatal implements grpclog.Logger and grpclog.LoggerV2.
func (l *Logger) Fatal(args ...interface{}) {
	l.fatal(l.log, args...)
}

// Fatalf implements grpclog.Logger and grpclog.LoggerV2.
func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.fatalf(l.log, format, args...)
}

// Fatalln implements grpclog.Logger and grpclog.LoggerV2.
func (l *Logger) Fatalln(args ...interface{}) {
	l.fatal(l.log, args...)
}

// Print implements grpclog.Logger.
func (l *Logger) Print(args ...interface{}) {
	l.print(l.log, args...)
}

// Printf implements grpclog.Logger.
func (l *Logger) Printf(format string, args ...interface{}) {
	l.printf(l.log, format, args...)
}

// Println implements grpclog.Logger.
func (l *Logger) Println(args ...interface{}) {
	l.print(l.log, args...)
}

// V implements grpclog.LoggerV2.
func (l *Logger) V(lvl int) bool {
	return lvl <= l.v
}
