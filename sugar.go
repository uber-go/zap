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
	"fmt"
	"time"
)

// A SugaredLogger wraps the core Logger functionality in a slower, but less
// verbose, API.
type SugaredLogger struct {
	core Logger
}

// Sugar converts a Logger to a SugaredLogger.
func Sugar(core Logger) *SugaredLogger {
	return &SugaredLogger{core}
}

// Desugar unwraps a SugaredLogger, exposing the original Logger.
func Desugar(s *SugaredLogger) Logger {
	return s.core
}

// With adds a variadic number of key-value pairs to the logging context.
// Even-indexed arguments are treated as keys, and are converted to strings
// (with fmt.Sprint) if necessary. The keys are then zipped with the
// odd-indexed values into typed fields, falling back to a reflection-based
// encoder if necessary.
//
// For example,
//   sugaredLogger.With(
//     "hello", "world",
//     "failure", errors.New("oh no"),
//     "count", 42,
//     "user", User{name: "alice"},
//  )
// is the equivalent of
//   coreLogger.With(
//     String("hello", "world"),
//     String("failure", "oh no"),
//     Int("count", 42),
//     Object("user", User{name: "alice"}),
//   )
func (s *SugaredLogger) With(args ...interface{}) *SugaredLogger {
	return &SugaredLogger{core: s.core.With(sweetenFields(args, s.core)...)}
}

// WithStack adds a complete stack trace to the logger's context, using the key
// "stacktrace".
func (s *SugaredLogger) WithStack() *SugaredLogger {
	return &SugaredLogger{core: s.core.With(Stack())}
}

// Debug logs a message and some key-value pairs at DebugLevel. Keys and values
// are treated as they are in the With method.
func (s *SugaredLogger) Debug(msg interface{}, keysAndValues ...interface{}) {
	s.core.Debug(sweetenMsg(msg), sweetenFields(keysAndValues, s.core)...)
}

// Debugf uses fmt.Sprintf to construct a dynamic message and logs it at
// DebugLevel. It doesn't add to the message's structured context.
func (s *SugaredLogger) Debugf(template string, args ...interface{}) {
	s.Debug(fmt.Sprintf(template, args...))
}

// Info logs a message and some key-value pairs at InfoLevel. Keys and values
// are treated as they are in the With method.
func (s *SugaredLogger) Info(msg interface{}, keysAndValues ...interface{}) {
	s.core.Info(sweetenMsg(msg), sweetenFields(keysAndValues, s.core)...)
}

// Infof uses fmt.Sprintf to construct a dynamic message and logs it at
// InfoLevel. It doesn't add to the message's structured context.
func (s *SugaredLogger) Infof(template string, args ...interface{}) {
	s.Info(fmt.Sprintf(template, args...))
}

// Warn logs a message and some key-value pairs at WarnLevel. Keys and values
// are treated as they are in the With method.
func (s *SugaredLogger) Warn(msg interface{}, keysAndValues ...interface{}) {
	s.core.Warn(sweetenMsg(msg), sweetenFields(keysAndValues, s.core)...)
}

// Warnf uses fmt.Sprintf to construct a dynamic message and logs it at
// WarnLevel. It doesn't add to the message's structured context.
func (s *SugaredLogger) Warnf(template string, args ...interface{}) {
	s.Warn(fmt.Sprintf(template, args...))
}

// Error logs a message and some key-value pairs at ErrorLevel. Keys and values
// are treated as they are in the With method.
func (s *SugaredLogger) Error(msg interface{}, keysAndValues ...interface{}) {
	s.core.Error(sweetenMsg(msg), sweetenFields(keysAndValues, s.core)...)
}

// Errorf uses fmt.Sprintf to construct a dynamic message and logs it at
// ErrorLevel. It doesn't add to the message's structured context.
func (s *SugaredLogger) Errorf(template string, args ...interface{}) {
	s.Error(fmt.Sprintf(template, args...))
}

// Panic logs a message and some key-value pairs at PanicLevel, then panics.
// Keys and values are treated as they are in the With method.
func (s *SugaredLogger) Panic(msg interface{}, keysAndValues ...interface{}) {
	s.core.Panic(sweetenMsg(msg), sweetenFields(keysAndValues, s.core)...)
}

// Panicf uses fmt.Sprintf to construct a dynamic message and logs it at
// PanicLevel, then panics. It doesn't add to the message's structured context.
func (s *SugaredLogger) Panicf(template string, args ...interface{}) {
	s.Panic(fmt.Sprintf(template, args...))
}

// Fatal logs a message and some key-value pairs at FatalLevel, then calls
// os.Exit(1). Keys and values are treated as they are in the With method.
func (s *SugaredLogger) Fatal(msg interface{}, keysAndValues ...interface{}) {
	s.core.Fatal(sweetenMsg(msg), sweetenFields(keysAndValues, s.core)...)
}

// Fatalf uses fmt.Sprintf to construct a dynamic message and logs it at
// FatalLevel, then calls os.Exit(1). It doesn't add to the message's
// structured context.
func (s *SugaredLogger) Fatalf(template string, args ...interface{}) {
	s.Fatal(fmt.Sprintf(template, args...))
}

// DFatal logs a message and some key-value pairs using the underlying logger's
// DFatal method. Keys and values are treated as they are in the With
// method. (See Logger.DFatal for details.)
func (s *SugaredLogger) DFatal(msg interface{}, keysAndValues ...interface{}) {
	s.core.DFatal(sweetenMsg(msg), sweetenFields(keysAndValues, s.core)...)
}

// DFatalf uses fmt.Sprintf to construct a dynamic message, which is passed to
// the underlying Logger's DFatal method. (See Logger.DFatal for details.) It
// doesn't add to the message's structured context.
func (s *SugaredLogger) DFatalf(template string, args ...interface{}) {
	s.DFatal(fmt.Sprintf(template, args...))
}

func sweetenFields(args []interface{}, errLogger Logger) []Field {
	if len(args) == 0 {
		return nil
	}
	if len(args)%2 == 1 {
		errLogger.DFatal(
			"Passed an odd number of keys and values to SugaredLogger, ignoring last.",
			Object("ignored", args[len(args)-1]),
		)
	}

	fields := make([]Field, len(args)/2)
	for i := range fields {
		key := sweetenMsg(args[2*i])

		switch val := args[2*i+1].(type) {
		case LogMarshaler:
			fields[i] = Marshaler(key, val)
		case bool:
			fields[i] = Bool(key, val)
		case float64:
			fields[i] = Float64(key, val)
		case float32:
			fields[i] = Float64(key, float64(val))
		case int:
			fields[i] = Int(key, val)
		case int64:
			fields[i] = Int64(key, val)
		case int32:
			fields[i] = Int64(key, int64(val))
		case int16:
			fields[i] = Int64(key, int64(val))
		case int8:
			fields[i] = Int64(key, int64(val))
		case uint:
			fields[i] = Uint(key, val)
		case uint64:
			fields[i] = Uint64(key, val)
		case uint32:
			fields[i] = Uint64(key, uint64(val))
		case uint16:
			fields[i] = Uint64(key, uint64(val))
		case uint8:
			fields[i] = Uint64(key, uint64(val))
		case uintptr:
			fields[i] = Uintptr(key, val)
		case string:
			fields[i] = String(key, val)
		case time.Time:
			fields[i] = Time(key, val)
		case time.Duration:
			fields[i] = Duration(key, val)
		case error:
			fields[i] = String(key, val.Error())
		case fmt.Stringer:
			fields[i] = String(key, val.String())
		default:
			fields[i] = Object(key, val)
		}
	}
	return fields
}

func sweetenMsg(msg interface{}) string {
	if m, ok := msg.(string); ok {
		return m
	}
	return fmt.Sprint(msg)
}
