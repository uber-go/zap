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

const (
	_oddNumberErrMsg    = "Passed an odd number of keys and values to SugaredLogger, ignoring last."
	_nonStringKeyErrMsg = "Passed a non-string key."
)

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
// Even-indexed arguments are treated as keys and zipped with the odd-indexed
// values using the Any field constructor.
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
//
// Note that the keys should be strings. In development, passing a non-string
// key panics. In production, the logger is more forgiving: a separate error is
// logged, but the key is coerced to a string with fmt.Sprint and execution
// continues. Passing an odd number of arguments triggers similar behavior:
// panics in development and errors in production.
func (s *SugaredLogger) With(args ...interface{}) *SugaredLogger {
	return s.WithFields(s.sweetenFields(args)...)
}

// WithFields adds structured fields to the logging context, much like the
// base Logger. It allows the sugared logger to use more specialized
// fields (like Stack).
func (s *SugaredLogger) WithFields(fs ...zapcore.Field) *SugaredLogger {
	return &SugaredLogger{core: s.core.With(fs...)}
}

// Debug uses fmt.Sprint to construct and log a message.
func (s *SugaredLogger) Debug(args ...interface{}) {
	s.log(DebugLevel, fmt.Sprint(args...), nil)
}

// Info uses fmt.Sprint to construct and log a message.
func (s *SugaredLogger) Info(args ...interface{}) {
	s.log(InfoLevel, fmt.Sprint(args...), nil)
}

// Warn uses fmt.Sprint to construct and log a message.
func (s *SugaredLogger) Warn(args ...interface{}) {
	s.log(WarnLevel, fmt.Sprint(args...), nil)
}

// Error uses fmt.Sprint to construct and log a message.
func (s *SugaredLogger) Error(args ...interface{}) {
	s.log(ErrorLevel, fmt.Sprint(args...), nil)
}

// DPanic uses fmt.Sprint to construct and log a message. In development, the
// logger then panics. (See DPanicLevel for details.)
func (s *SugaredLogger) DPanic(args ...interface{}) {
	s.log(DPanicLevel, fmt.Sprint(args...), nil)
}

// Panic uses fmt.Sprint to construct and log a message, then panics.
func (s *SugaredLogger) Panic(args ...interface{}) {
	s.log(PanicLevel, fmt.Sprint(args...), nil)
}

// Fatal uses fmt.Sprint to construct and log a message, then calls os.Exit.
func (s *SugaredLogger) Fatal(args ...interface{}) {
	s.log(FatalLevel, fmt.Sprint(args...), nil)
}

// Debugf uses fmt.Sprintf to log a templated message.
func (s *SugaredLogger) Debugf(template string, args ...interface{}) {
	s.log(DebugLevel, fmt.Sprintf(template, args...), nil)
}

// Infof uses fmt.Sprintf to log a templated message.
func (s *SugaredLogger) Infof(template string, args ...interface{}) {
	s.log(InfoLevel, fmt.Sprintf(template, args...), nil)
}

// Warnf uses fmt.Sprintf to log a templated message.
func (s *SugaredLogger) Warnf(template string, args ...interface{}) {
	s.log(WarnLevel, fmt.Sprintf(template, args...), nil)
}

// Errorf uses fmt.Sprintf to log a templated message.
func (s *SugaredLogger) Errorf(template string, args ...interface{}) {
	s.log(ErrorLevel, fmt.Sprintf(template, args...), nil)
}

// DPanicf uses fmt.Sprintf to log a templated message. In development, the
// logger then panics. (See DPanicLevel for details.)
func (s *SugaredLogger) DPanicf(template string, args ...interface{}) {
	s.log(DPanicLevel, fmt.Sprintf(template, args...), nil)
}

// Panicf uses fmt.Sprintf to log a templated message, then panics.
func (s *SugaredLogger) Panicf(template string, args ...interface{}) {
	s.log(PanicLevel, fmt.Sprintf(template, args...), nil)
}

// Fatalf uses fmt.Sprintf to log a templated message, then calls os.Exit.
func (s *SugaredLogger) Fatalf(template string, args ...interface{}) {
	s.log(FatalLevel, fmt.Sprintf(template, args...), nil)
}

// Debugw logs a message with some additional context. The variadic key-value
// pairs are treated as they are in With.
//
// When debug-level logging is disabled, this is much faster than
//  s.With(keysAndValues).Debug(msg)
func (s *SugaredLogger) Debugw(msg string, keysAndValues ...interface{}) {
	s.log(DebugLevel, msg, keysAndValues)
}

// Infow logs a message with some additional context. The variadic key-value
// pairs are treated as they are in With.
func (s *SugaredLogger) Infow(msg string, keysAndValues ...interface{}) {
	s.log(InfoLevel, msg, keysAndValues)
}

// Warnw logs a message with some additional context. The variadic key-value
// pairs are treated as they are in With.
func (s *SugaredLogger) Warnw(msg string, keysAndValues ...interface{}) {
	s.log(WarnLevel, msg, keysAndValues)
}

// Errorw logs a message with some additional context. The variadic key-value
// pairs are treated as they are in With.
func (s *SugaredLogger) Errorw(msg string, keysAndValues ...interface{}) {
	s.log(ErrorLevel, msg, keysAndValues)
}

// DPanicw logs a message with some additional context. In development, the
// logger then panics. (See DPanicLevel for details.) The variadic key-value
// pairs are treated as they are in With.
func (s *SugaredLogger) DPanicw(msg string, keysAndValues ...interface{}) {
	s.log(DPanicLevel, msg, keysAndValues)
}

// Panicw logs a message with some additional context, then panics. The
// variadic key-value pairs are treated as they are in With.
func (s *SugaredLogger) Panicw(msg string, keysAndValues ...interface{}) {
	s.log(PanicLevel, msg, keysAndValues)
}

// Fatalw logs a message with some additional context, then calls os.Exit. The
// variadic key-value pairs are treated as they are in With.
func (s *SugaredLogger) Fatalw(msg string, keysAndValues ...interface{}) {
	s.log(FatalLevel, msg, keysAndValues)
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
		keyIdx := 2 * i
		val := args[keyIdx+1]
		key, ok := args[keyIdx].(string)
		if !ok {
			s.core.DPanic(
				_nonStringKeyErrMsg,
				Int("position", keyIdx),
				Any("key", args[keyIdx]),
				Any("value", val),
			)
			key = fmt.Sprint(args[keyIdx])
		}
		fields[i] = Any(key, val)
	}
	return fields
}
