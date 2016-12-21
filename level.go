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
	"errors"
	"fmt"

	"github.com/uber-go/atomic"
)

var errMarshalNilLevel = errors.New("can't marshal a nil *Level to text")

// A Level is a logging priority. Higher levels are more important.
//
// Note that Level satisfies the Option interface, so any Level can be passed to
// New to override the default logging priority.
type Level int32

// LevelEnabler decides whether a given logging level is enabled when logging a
// message.
//
// Enablers are intended to be used to implement deterministic filters;
// concerns like sampling are better implemented as a Logger implementation.
//
// Each concrete Level value implements a static LevelEnabler which returns
// true for itself and all higher logging levels. For example WarnLevel.Enabled()
// will return true for WarnLevel, ErrorLevel, PanicLevel, and FatalLevel, but
// return false for InfoLevel and DebugLevel.
type LevelEnabler interface {
	Enabled(Level) bool
}

const (
	invalidLevel Level = iota - 2

	// DebugLevel logs are typically voluminous, and are usually disabled in
	// production.
	DebugLevel
	// InfoLevel is the default logging priority.
	InfoLevel
	// WarnLevel logs are more important than Info, but don't need individual
	// human review.
	WarnLevel
	// ErrorLevel logs are high-priority. If an application is running smoothly,
	// it shouldn't generate any error-level logs.
	ErrorLevel
	// DPanicLevel logs are particularly important errors. In development the
	// logger panics after writing the message.
	DPanicLevel
	// PanicLevel logs a message, then panics.
	PanicLevel
	// FatalLevel logs a message, then calls os.Exit(1).
	FatalLevel
)

// LevelEnablerFunc is a convenient way to implement LevelEnabler around an
// anonymous function. It is also a valid Option to pass to a logger.
type LevelEnablerFunc func(Level) bool

// This allows an LevelEnablerFunc to be used as an option.
func (f LevelEnablerFunc) apply(m *Meta) { m.LevelEnabler = f }

// Enabled calls the wrapped function.
func (f LevelEnablerFunc) Enabled(lvl Level) bool { return f(lvl) }

// String returns a lower-case ASCII representation of the log level.
func (l Level) String() string {
	switch l {
	case DebugLevel:
		return "debug"
	case InfoLevel:
		return "info"
	case WarnLevel:
		return "warn"
	case ErrorLevel:
		return "error"
	case DPanicLevel:
		return "dpanic"
	case PanicLevel:
		return "panic"
	case FatalLevel:
		return "fatal"
	default:
		return fmt.Sprintf("Level(%d)", l)
	}
}

// MarshalText marshals the Level to text. Note that the text representation
// drops the -Level suffix (see example).
func (l *Level) MarshalText() ([]byte, error) {
	if l == nil {
		return nil, errMarshalNilLevel
	}
	return []byte(l.String()), nil
}

// UnmarshalText unmarshals text to a level. Like MarshalText, UnmarshalText
// expects the text representation of a Level to drop the -Level suffix (see
// example).
//
// In particular, this makes it easy to configure logging levels using YAML,
// TOML, or JSON files.
func (l *Level) UnmarshalText(text []byte) error {
	switch string(text) {
	case "debug":
		*l = DebugLevel
	case "info":
		*l = InfoLevel
	case "warn":
		*l = WarnLevel
	case "error":
		*l = ErrorLevel
	case "dpanic":
		*l = DPanicLevel
	case "panic":
		*l = PanicLevel
	case "fatal":
		*l = FatalLevel
	default:
		return fmt.Errorf("unrecognized level: %v", string(text))
	}
	return nil
}

// Set sets the level for the flag.Value interface.
func (l *Level) Set(s string) error {
	switch s {
	case "debug":
		*l = DebugLevel
	case "info":
		*l = InfoLevel
	case "warn":
		*l = WarnLevel
	case "error":
		*l = ErrorLevel
	case "dpanic":
		*l = DPanicLevel
	case "panic":
		*l = PanicLevel
	case "fatal":
		*l = FatalLevel
	default:
		return fmt.Errorf("unrecognized level: %q", s)
	}
	return nil
}

// Get gets the level for the flag.Getter interface.
func (l *Level) Get() interface{} {
	return *l
}

// Enabled returns true if the given level is at or above this level.
func (l Level) Enabled(lvl Level) bool {
	return lvl >= l
}

// DynamicLevel creates an atomically changeable dynamic logging level.  The
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
func (lvl AtomicLevel) Enabled(l Level) bool {
	return lvl.Level().Enabled(l)
}

// Level returns the minimum enabled log level.
func (lvl AtomicLevel) Level() Level {
	return Level(lvl.l.Load())
}

// SetLevel alters the logging level.
func (lvl AtomicLevel) SetLevel(l Level) {
	lvl.l.Store(int32(l))
}
