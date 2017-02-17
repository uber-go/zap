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

/*
Package syslog implements syslog functionality for zap.
*/
package syslog

import (
	"log/syslog"

	"go.uber.org/zap/zapcore"
)

const (
	EmergSeverity Severity = iota
	AlertSeverity
	CritSeverity
	ErrSeverity
	WarningSeverity
	NoticeSeverity
	InfoSeverity
	DebugSeverity
)

// Severity is a syslog severity.
type Severity int

// Logger logs to Syslog.
type Logger interface {
	Write(severity Severity, message string) error
	Close() error
}

// NewLogger returns a new Logger that wraps the standard syslog.Logger.
//
// Note the log/syslog package is frozen now, and we may want to switch
// to another potential implementation in the future (hence the interface).
func NewLogger(delegate *syslog.Writer) Logger {
	return newLogger(delegate)
}

// Pusher pushes log entries.
type Pusher interface {
	Push(level zapcore.Level, data []byte) error
	Sync() error
}

// NewPusher returns a new Pusher for the given Logger.
func NewPusher(logger Logger) Pusher {
	return newPusher(logger)
}
