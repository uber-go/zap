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

import "go.uber.org/zap"

// LoggerOption is an option for a new Logger.
type LoggerOption func(*Logger)

// WithDebug says to print the debug level instead of the default info level.
func WithDebug() LoggerOption {
	return func(logger *Logger) {
		logger.printFunc = (*zap.SugaredLogger).Debug
		logger.printfFunc = (*zap.SugaredLogger).Debugf
	}
}

// NewLogger returns a new Logger.
func NewLogger(l *zap.Logger, options ...LoggerOption) *Logger {
	logger := &Logger{
		l.Sugar(),
		(*zap.SugaredLogger).Fatal,
		(*zap.SugaredLogger).Fatalf,
		(*zap.SugaredLogger).Info,
		(*zap.SugaredLogger).Infof,
	}
	for _, option := range options {
		option(logger)
	}
	return logger
}

// Logger is a logger that is compatible with the grpclog Logger interface.
type Logger struct {
	sugaredLogger *zap.SugaredLogger
	fatalFunc     func(*zap.SugaredLogger, ...interface{})
	fatalfFunc    func(*zap.SugaredLogger, string, ...interface{})
	printFunc     func(*zap.SugaredLogger, ...interface{})
	printfFunc    func(*zap.SugaredLogger, string, ...interface{})
}

// Fatal implements grpclog.Logger#Fatal.
func (l *Logger) Fatal(args ...interface{}) {
	l.fatalFunc(l.sugaredLogger, args...)
}

// Fatalf implements grpclog.Logger#Fatalf.
func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.fatalfFunc(l.sugaredLogger, format, args...)
}

// Fatalln implements grpclog.Logger#Fatalln.
func (l *Logger) Fatalln(args ...interface{}) {
	l.fatalFunc(l.sugaredLogger, args...)
}

// Print implements grpclog.Logger#Print.
func (l *Logger) Print(args ...interface{}) {
	l.printFunc(l.sugaredLogger, args...)
}

// Printf implements grpclog.Logger#Printf.
func (l *Logger) Printf(format string, args ...interface{}) {
	l.printfFunc(l.sugaredLogger, format, args...)
}

// Println implements grpclog.Logger#Println.
func (l *Logger) Println(args ...interface{}) {
	l.printFunc(l.sugaredLogger, args...)
}
