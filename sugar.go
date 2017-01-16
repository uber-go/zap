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

	"go.uber.org/zap/zapcore"
)

const _oddNumberErrMsg = "Passed an odd number of keys and values to SugaredLogger, ignoring last."

// A SugaredLogger wraps the core Logger functionality in a slower, but less
// verbose, API.
type SugaredLogger struct {
	core Logger
}

// Sugar converts a Logger to a SugaredLogger.
func Sugar(core Logger) *SugaredLogger {
	// TODO: increment caller skip.
	return &SugaredLogger{core}
}

// Desugar unwraps a SugaredLogger, exposing the original Logger.
func Desugar(s *SugaredLogger) Logger {
	// TODO: decrement caller skip.
	return s.core
}

// With adds a variadic number of key-value pairs to the logging context.
// Even-indexed arguments are treated as keys, and are converted to strings
// (with fmt.Sprint) if necessary. The keys are then zipped with the
// odd-indexed values into typed fields using the Any field constructor.
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
	return &SugaredLogger{core: s.core.With(s.sweetenFields(args)...)}
}

// Debug logs a message and some key-value pairs at DebugLevel. Keys and values
// are treated as they are in the With method.
func (s *SugaredLogger) Debug(msg interface{}, keysAndValues ...interface{}) {
	s.log(DebugLevel, sweetenMsg(msg), keysAndValues)
}

// Debugf uses fmt.Sprintf to construct a dynamic message and logs it at
// DebugLevel. It doesn't add to the message's structured context.
func (s *SugaredLogger) Debugf(template string, args ...interface{}) {
	s.log(DebugLevel, fmt.Sprintf(template, args...), nil)
}

// Info logs a message and some key-value pairs at InfoLevel. Keys and values
// are treated as they are in the With method.
func (s *SugaredLogger) Info(msg interface{}, keysAndValues ...interface{}) {
	s.log(InfoLevel, sweetenMsg(msg), keysAndValues)
}

// Infof uses fmt.Sprintf to construct a dynamic message and logs it at
// InfoLevel. It doesn't add to the message's structured context.
func (s *SugaredLogger) Infof(template string, args ...interface{}) {
	s.log(InfoLevel, fmt.Sprintf(template, args...), nil)
}

// Warn logs a message and some key-value pairs at WarnLevel. Keys and values
// are treated as they are in the With method.
func (s *SugaredLogger) Warn(msg interface{}, keysAndValues ...interface{}) {
	s.log(WarnLevel, sweetenMsg(msg), keysAndValues)
}

// Warnf uses fmt.Sprintf to construct a dynamic message and logs it at
// WarnLevel. It doesn't add to the message's structured context.
func (s *SugaredLogger) Warnf(template string, args ...interface{}) {
	s.log(WarnLevel, fmt.Sprintf(template, args...), nil)
}

// Error logs a message and some key-value pairs at ErrorLevel. Keys and values
// are treated as they are in the With method.
func (s *SugaredLogger) Error(msg interface{}, keysAndValues ...interface{}) {
	s.log(ErrorLevel, sweetenMsg(msg), keysAndValues)
}

// Errorf uses fmt.Sprintf to construct a dynamic message and logs it at
// ErrorLevel. It doesn't add to the message's structured context.
func (s *SugaredLogger) Errorf(template string, args ...interface{}) {
	s.log(ErrorLevel, fmt.Sprintf(template, args...), nil)
}

// DPanic logs a message and some key-value pairs using the underlying logger's
// DPanic method. Keys and values are treated as they are in the With
// method. (See Logger.DPanic for details.)
func (s *SugaredLogger) DPanic(msg interface{}, keysAndValues ...interface{}) {
	s.log(DPanicLevel, sweetenMsg(msg), keysAndValues)
}

// DPanicf uses fmt.Sprintf to construct a dynamic message, which is passed to
// the underlying Logger's DPanic method. (See Logger.DPanic for details.) It
// doesn't add to the message's structured context.
func (s *SugaredLogger) DPanicf(template string, args ...interface{}) {
	s.log(DPanicLevel, fmt.Sprintf(template, args...), nil)
}

// Panic logs a message and some key-value pairs at PanicLevel, then panics.
// Keys and values are treated as they are in the With method.
func (s *SugaredLogger) Panic(msg interface{}, keysAndValues ...interface{}) {
	s.log(PanicLevel, sweetenMsg(msg), keysAndValues)
}

// Panicf uses fmt.Sprintf to construct a dynamic message and logs it at
// PanicLevel, then panics. It doesn't add to the message's structured context.
func (s *SugaredLogger) Panicf(template string, args ...interface{}) {
	s.log(PanicLevel, fmt.Sprintf(template, args...), nil)
}

// Fatal logs a message and some key-value pairs at FatalLevel, then calls
// os.Exit(1). Keys and values are treated as they are in the With method.
func (s *SugaredLogger) Fatal(msg interface{}, keysAndValues ...interface{}) {
	s.log(FatalLevel, sweetenMsg(msg), keysAndValues)
}

// Fatalf uses fmt.Sprintf to construct a dynamic message and logs it at
// FatalLevel, then calls os.Exit(1). It doesn't add to the message's
// structured context.
func (s *SugaredLogger) Fatalf(template string, args ...interface{}) {
	s.log(FatalLevel, fmt.Sprintf(template, args...), nil)
}

func (s *SugaredLogger) log(lvl zapcore.Level, msg string, context []interface{}) {
	if ce := s.core.Check(lvl, msg); ce != nil {
		ce.Write(s.sweetenFields(context)...)
	}
}

func (s *SugaredLogger) sweetenFields(args []interface{}) []zapcore.Field {
	if len(args) == 0 {
		return nil
	}
	if len(args)%2 == 1 {
		s.core.DPanic(_oddNumberErrMsg, Any("ignored", args[len(args)-1]))
	}

	fields := make([]zapcore.Field, len(args)/2)
	for i := range fields {
		key := sweetenMsg(args[2*i])
		fields[i] = Any(key, args[2*i+1])
	}
	return fields
}

func sweetenMsg(msg interface{}) string {
	if str, ok := msg.(string); ok {
		return str
	}
	return fmt.Sprint(msg)
}
