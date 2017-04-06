// Copyright (c) 2017 Uber Technologies, Inc.
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

// Logger mimics grpclog's Logger interface.
type Logger interface {
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
	Fatalln(args ...interface{})
	Print(args ...interface{})
	Printf(format string, args ...interface{})
	Println(args ...interface{})
}

// NewLogger returns a new Logger that wraps the given *zap.Logger.
//
// Print statements will log at the Info level.
func NewLogger(l *zap.Logger) Logger {
	return &logger{l.Sugar()}
}

type logger struct {
	*zap.SugaredLogger
}

func (l *logger) Fatalln(args ...interface{}) {
	l.SugaredLogger.Fatal(args...)
}

func (l *logger) Print(args ...interface{}) {
	l.SugaredLogger.Info(args...)
}

func (l *logger) Printf(format string, args ...interface{}) {
	l.SugaredLogger.Infof(format, args...)
}

func (l *logger) Println(args ...interface{}) {
	l.SugaredLogger.Info(args...)
}
