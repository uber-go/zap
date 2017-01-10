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

package zap

import (
	"go.uber.org/atomic"
	"go.uber.org/zap/zapcore"
)

const (
	// Re-export the named levels for user convenience.

	// DebugLevel logs are typically voluminous, and are usually disabled in
	// production.
	DebugLevel = zapcore.DebugLevel
	// InfoLevel is the default logging priority.
	InfoLevel = zapcore.InfoLevel
	// WarnLevel logs are more important than Info, but don't need individual
	// human review.
	WarnLevel = zapcore.WarnLevel
	// ErrorLevel logs are high-priority. If an application is running smoothly,
	// it shouldn't generate any error-level logs.
	ErrorLevel = zapcore.ErrorLevel
	// DPanicLevel logs are particularly important errors. In development the
	// logger panics after writing the message.
	DPanicLevel = zapcore.DPanicLevel
	// PanicLevel logs a message, then panics.
	PanicLevel = zapcore.PanicLevel
	// FatalLevel logs a message, then calls os.Exit(1).
	FatalLevel = zapcore.FatalLevel
)

// LevelEnablerFunc is a convenient way to implement LevelEnabler around an
// anonymous function. It is also a valid Option to pass to a logger.
type LevelEnablerFunc func(zapcore.Level) bool

// Enabled calls the wrapped function.
func (f LevelEnablerFunc) Enabled(lvl zapcore.Level) bool { return f(lvl) }

// DynamicLevel creates an atomically changeable, dynamic logging level. The
// returned level can be passed as a logger option just like a concrete level.
//
// The value's SetLevel() method may be called later to change the enabled
// logging level of all loggers that were passed the value (either explicitly,
// or by creating sub-loggers with Logger.With).
func DynamicLevel() AtomicLevel {
	return AtomicLevel{
		l: atomic.NewInt32(int32(InfoLevel)),
	}
}

// AtomicLevel wraps an atomically change-able Level value. It must be created
// by the DynamicLevel() function to allocate the internal atomic pointer.
type AtomicLevel struct {
	l *atomic.Int32
}

// Enabled loads the level value, and calls its Enabled method.
func (lvl AtomicLevel) Enabled(l zapcore.Level) bool {
	return lvl.Level().Enabled(l)
}

// Level returns the minimum enabled log level.
func (lvl AtomicLevel) Level() zapcore.Level {
	return zapcore.Level(int8(lvl.l.Load()))
}

// SetLevel alters the logging level.
func (lvl AtomicLevel) SetLevel(l zapcore.Level) {
	lvl.l.Store(int32(l))
}
