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

// A Ctx is a loosely-typed collection of structured log fields. It's a simple
// alias to reduce the wordiness of log sites.
type Ctx map[string]interface{}

// A SugaredLogger wraps the core Logger functionality in a slightly slower,
// but less verbose, API.
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

// With adds a loosely-typed collection of key-value pairs to the logging
// context.
//
// For example,
//   sugaredLogger.With(zap.Ctx{
//     "hello": "world",
//     "count": 42,
//     "user": User{name: "alice"},
//  })
// is equivalent to
//   logger.With(
//     String("hello", "world"),
//     Int("count", 42),
//     Any("user", User{name: "alice"}),
//   )
func (s *SugaredLogger) With(fields Ctx) *SugaredLogger {
	return &SugaredLogger{core: s.core.With(s.sweeten(fields)...)}
}

// Debug logs a message, along with any context accumulated on the logger, at
// DebugLevel.
func (s *SugaredLogger) Debug(msg string) {
	s.log(DebugLevel, msg, nil)
}

// DebugWith adds to the logger's context, then logs a message and the combined
// context at DebugLevel. For context that's only relevant at one log site,
// it's faster than logger.With(ctx).Debug(msg).
func (s *SugaredLogger) DebugWith(msg string, fields Ctx) {
	s.log(DebugLevel, msg, fields)
}

// Debugf uses fmt.Sprintf to construct a templated message, then passes it to
// Debug.
func (s *SugaredLogger) Debugf(template string, args ...interface{}) {
	s.log(DebugLevel, fmt.Sprintf(template, args...), nil)
}

// Info logs a message, along with any context accumulated on the logger, at
// InfoLevel.
func (s *SugaredLogger) Info(msg string) {
	s.log(InfoLevel, msg, nil)
}

// InfoWith adds to the logger's context, then logs a message and the combined
// context at InfoLevel. For context that's only relevant at one log site,
// it's faster than logger.With(ctx).Info(msg).
func (s *SugaredLogger) InfoWith(msg string, fields Ctx) {
	s.log(InfoLevel, msg, fields)
}

// Infof uses fmt.Sprintf to construct a templated message, then passes it to
// Info.
func (s *SugaredLogger) Infof(template string, args ...interface{}) {
	s.log(InfoLevel, fmt.Sprintf(template, args...), nil)
}

// Warn logs a message, along with any context accumulated on the logger, at
// WarnLevel.
func (s *SugaredLogger) Warn(msg string) {
	s.log(WarnLevel, msg, nil)
}

// WarnWith adds to the logger's context, then logs a message and the combined
// context at WarnLevel. For context that's only relevant at one log site,
// it's faster than logger.With(ctx).Warn(msg).
func (s *SugaredLogger) WarnWith(msg string, fields Ctx) {
	s.log(WarnLevel, msg, fields)
}

// Warnf uses fmt.Sprintf to construct a templated message, then passes it to
// Warn.
func (s *SugaredLogger) Warnf(template string, args ...interface{}) {
	s.log(WarnLevel, fmt.Sprintf(template, args...), nil)
}

// Error logs a message, along with any context accumulated on the logger, at
// ErrorLevel.
func (s *SugaredLogger) Error(msg string) {
	s.log(ErrorLevel, msg, nil)
}

// ErrorWith adds to the logger's context, then logs a message and the combined
// context at ErrorLevel. For context that's only relevant at one log site,
// it's faster than logger.With(ctx).Error(msg).
func (s *SugaredLogger) ErrorWith(msg string, fields Ctx) {
	s.log(ErrorLevel, msg, fields)
}

// Errorf uses fmt.Sprintf to construct a templated message, then passes it to
// Error.
func (s *SugaredLogger) Errorf(template string, args ...interface{}) {
	s.log(ErrorLevel, fmt.Sprintf(template, args...), nil)
}

// DPanic logs a message, along with any context accumulated on the logger, at
// DPanicLevel.
func (s *SugaredLogger) DPanic(msg string) {
	s.log(DPanicLevel, msg, nil)
}

// DPanicWith adds to the logger's context, then logs a message and the combined
// context at DPanicLevel. For context that's only relevant at one log site,
// it's faster than logger.With(ctx).DPanic(msg).
func (s *SugaredLogger) DPanicWith(msg string, fields Ctx) {
	s.log(DPanicLevel, msg, fields)
}

// DPanicf uses fmt.Sprintf to construct a templated message, then passes it to
// DPanic.
func (s *SugaredLogger) DPanicf(template string, args ...interface{}) {
	s.log(DPanicLevel, fmt.Sprintf(template, args...), nil)
}

// Panic logs a message, along with any context accumulated on the logger, at
// PanicLevel.
func (s *SugaredLogger) Panic(msg string) {
	s.log(PanicLevel, msg, nil)
}

// PanicWith adds to the logger's context, then logs a message and the combined
// context at PanicLevel. For context that's only relevant at one log site,
// it's faster than logger.With(ctx).Panic(msg).
func (s *SugaredLogger) PanicWith(msg string, fields Ctx) {
	s.log(PanicLevel, msg, fields)
}

// Panicf uses fmt.Sprintf to construct a templated message, then passes it to
// Panic.
func (s *SugaredLogger) Panicf(template string, args ...interface{}) {
	s.log(PanicLevel, fmt.Sprintf(template, args...), nil)
}

// Fatal logs a message, along with any context accumulated on the logger, at
// FatalLevel.
func (s *SugaredLogger) Fatal(msg string) {
	s.log(FatalLevel, msg, nil)
}

// FatalWith adds to the logger's context, then logs a message and the combined
// context at FatalLevel. For context that's only relevant at one log site,
// it's faster than logger.With(ctx).Fatal(msg).
func (s *SugaredLogger) FatalWith(msg string, fields Ctx) {
	s.log(FatalLevel, msg, fields)
}

// Fatalf uses fmt.Sprintf to construct a templated message, then passes it to
// Fatal.
func (s *SugaredLogger) Fatalf(template string, args ...interface{}) {
	s.log(FatalLevel, fmt.Sprintf(template, args...), nil)
}

func (s *SugaredLogger) log(lvl zapcore.Level, msg string, fields Ctx) {
	if ce := s.core.Check(lvl, msg); ce != nil {
		ce.Write(s.sweeten(fields)...)
	}
}

func (s *SugaredLogger) sweeten(fields Ctx) []zapcore.Field {
	if len(fields) == 0 {
		return nil
	}
	fs := make([]zapcore.Field, 0, len(fields))
	for key := range fields {
		fs = append(fs, Any(key, fields[key]))
	}
	return fs
}
