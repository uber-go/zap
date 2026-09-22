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
	"context"
	"fmt"

	"go.uber.org/zap/zapcore"

	"go.uber.org/multierr"
)

const (
	_oddNumberErrMsg    = "Ignored key without a value."
	_nonStringKeyErrMsg = "Ignored key-value pairs with non-string keys."
	_multipleErrMsg     = "Multiple errors without a key."
)

// A SugaredLogger wraps the base Logger functionality in a slower, but less
// verbose, API. Any Logger can be converted to a SugaredLogger with its Sugar
// method.
//
// Unlike the Logger, the SugaredLogger doesn't insist on structured logging.
// For each log level, it exposes four methods:
//
//   - methods named after the log level for log.Print-style logging
//   - methods ending in "w" for loosely-typed structured logging
//   - methods ending in "f" for log.Printf-style logging
//   - methods ending in "ln" for log.Println-style logging
//
// For example, the methods for InfoLevel are:
//
//	Info(...any)           Print-style logging
//	Infow(...any)          Structured logging (read as "info with")
//	Infof(string, ...any)  Printf-style logging
//	Infoln(...any)         Println-style logging
type SugaredLogger struct {
	base *Logger
}

// Desugar unwraps a SugaredLogger, exposing the original Logger. Desugaring
// is quite inexpensive, so it's reasonable for a single application to use
// both Loggers and SugaredLoggers, converting between them on the boundaries
// of performance-sensitive code.
func (s *SugaredLogger) Desugar() *Logger {
	base := s.base.clone()
	base.callerSkip -= 2
	return base
}

// Named adds a sub-scope to the logger's name. See Logger.Named for details.
func (s *SugaredLogger) Named(name string) *SugaredLogger {
	return &SugaredLogger{base: s.base.Named(name)}
}

// WithOptions clones the current SugaredLogger, applies the supplied Options,
// and returns the result. It's safe to use concurrently.
func (s *SugaredLogger) WithOptions(opts ...Option) *SugaredLogger {
	base := s.base.clone()
	for _, opt := range opts {
		opt.apply(base)
	}
	return &SugaredLogger{base: base}
}

// With adds a variadic number of fields to the logging context. It accepts a
// mix of strongly-typed Field objects and loosely-typed key-value pairs. When
// processing pairs, the first element of the pair is used as the field key
// and the second as the field value.
//
// For example,
//
//	 sugaredLogger.With(
//	   "hello", "world",
//	   "failure", errors.New("oh no"),
//	   Stack(),
//	   "count", 42,
//	   "user", User{Name: "alice"},
//	)
//
// is the equivalent of
//
//	unsugared.With(
//	  String("hello", "world"),
//	  String("failure", "oh no"),
//	  Stack(),
//	  Int("count", 42),
//	  Object("user", User{Name: "alice"}),
//	)
//
// Note that the keys in key-value pairs should be strings. In development,
// passing a non-string key panics. In production, the logger is more
// forgiving: a separate error is logged, but the key-value pair is skipped
// and execution continues. Passing an orphaned key triggers similar behavior:
// panics in development and errors in production.
func (s *SugaredLogger) With(args ...interface{}) *SugaredLogger {
	return &SugaredLogger{base: s.base.With(s.sweetenFields(args)...)}
}

// WithLazy adds a variadic number of fields to the logging context lazily.
// The fields are evaluated only if the logger is further chained with [With]
// or is written to with any of the log level methods.
// Until that occurs, the logger may retain references to objects inside the fields,
// and logging will reflect the state of an object at the time of logging,
// not the time of WithLazy().
//
// Similar to [With], fields added to the child don't affect the parent,
// and vice versa. Also, the keys in key-value pairs should be strings. In development,
// passing a non-string key panics, while in production it logs an error and skips the pair.
// Passing an orphaned key has the same behavior.
func (s *SugaredLogger) WithLazy(args ...interface{}) *SugaredLogger {
	return &SugaredLogger{base: s.base.WithLazy(s.sweetenFields(args)...)}
}

// Level reports the minimum enabled level for this logger.
//
// For NopLoggers, this is [zapcore.InvalidLevel].
func (s *SugaredLogger) Level() zapcore.Level {
	return zapcore.LevelOf(s.base.core)
}

// Log logs the provided arguments at provided level.
// Spaces are added between arguments when neither is a string.
func (s *SugaredLogger) Log(lvl zapcore.Level, args ...interface{}) {
	s.log(lvl, "", args, nil)
}

// LogCtx logs the provided arguments at provided level. Any fields extracted
// from ctx by the Context option are added to the entry.
// Spaces are added between arguments when neither is a string.
func (s *SugaredLogger) LogCtx(ctx context.Context, lvl zapcore.Level, args ...interface{}) {
	s.log(lvl, "", args, s.ctxFields(ctx))
}

// Debug logs the provided arguments at [DebugLevel].
// Spaces are added between arguments when neither is a string.
func (s *SugaredLogger) Debug(args ...interface{}) {
	s.log(DebugLevel, "", args, nil)
}

// DebugCtx logs the provided arguments at [DebugLevel]. Any fields extracted
// from ctx by the Context option are added to the entry.
// Spaces are added between arguments when neither is a string.
func (s *SugaredLogger) DebugCtx(ctx context.Context, args ...interface{}) {
	s.log(DebugLevel, "", args, s.ctxFields(ctx))
}

// Info logs the provided arguments at [InfoLevel].
// Spaces are added between arguments when neither is a string.
func (s *SugaredLogger) Info(args ...interface{}) {
	s.log(InfoLevel, "", args, nil)
}

// InfoCtx logs the provided arguments at [InfoLevel]. Any fields extracted
// from ctx by the Context option are added to the entry.
// Spaces are added between arguments when neither is a string.
func (s *SugaredLogger) InfoCtx(ctx context.Context, args ...interface{}) {
	s.log(InfoLevel, "", args, s.ctxFields(ctx))
}

// Warn logs the provided arguments at [WarnLevel].
// Spaces are added between arguments when neither is a string.
func (s *SugaredLogger) Warn(args ...interface{}) {
	s.log(WarnLevel, "", args, nil)
}

// WarnCtx logs the provided arguments at [WarnLevel]. Any fields extracted
// from ctx by the Context option are added to the entry.
// Spaces are added between arguments when neither is a string.
func (s *SugaredLogger) WarnCtx(ctx context.Context, args ...interface{}) {
	s.log(WarnLevel, "", args, s.ctxFields(ctx))
}

// Error logs the provided arguments at [ErrorLevel].
// Spaces are added between arguments when neither is a string.
func (s *SugaredLogger) Error(args ...interface{}) {
	s.log(ErrorLevel, "", args, nil)
}

// ErrorCtx logs the provided arguments at [ErrorLevel]. Any fields extracted
// from ctx by the Context option are added to the entry.
// Spaces are added between arguments when neither is a string.
func (s *SugaredLogger) ErrorCtx(ctx context.Context, args ...interface{}) {
	s.log(ErrorLevel, "", args, s.ctxFields(ctx))
}

// DPanic logs the provided arguments at [DPanicLevel].
// In development, the logger then panics. (See [DPanicLevel] for details.)
// Spaces are added between arguments when neither is a string.
func (s *SugaredLogger) DPanic(args ...interface{}) {
	s.log(DPanicLevel, "", args, nil)
}

// DPanicCtx logs the provided arguments at [DPanicLevel]. Any fields extracted
// from ctx by the Context option are added to the entry.
// In development, the logger then panics. (See [DPanicLevel] for details.)
// Spaces are added between arguments when neither is a string.
func (s *SugaredLogger) DPanicCtx(ctx context.Context, args ...interface{}) {
	s.log(DPanicLevel, "", args, s.ctxFields(ctx))
}

// Panic constructs a message with the provided arguments and panics.
// Spaces are added between arguments when neither is a string.
func (s *SugaredLogger) Panic(args ...interface{}) {
	s.log(PanicLevel, "", args, nil)
}

// PanicCtx constructs a message with the provided arguments and panics. Any
// fields extracted from ctx by the Context option are added to the entry.
// Spaces are added between arguments when neither is a string.
func (s *SugaredLogger) PanicCtx(ctx context.Context, args ...interface{}) {
	s.log(PanicLevel, "", args, s.ctxFields(ctx))
}

// Fatal constructs a message with the provided arguments and calls os.Exit.
// Spaces are added between arguments when neither is a string.
func (s *SugaredLogger) Fatal(args ...interface{}) {
	s.log(FatalLevel, "", args, nil)
}

// FatalCtx constructs a message with the provided arguments and calls os.Exit.
// Any fields extracted from ctx by the Context option are added to the entry.
// Spaces are added between arguments when neither is a string.
func (s *SugaredLogger) FatalCtx(ctx context.Context, args ...interface{}) {
	s.log(FatalLevel, "", args, s.ctxFields(ctx))
}

// Logf formats the message according to the format specifier
// and logs it at provided level.
func (s *SugaredLogger) Logf(lvl zapcore.Level, template string, args ...interface{}) {
	s.log(lvl, template, args, nil)
}

// LogfCtx formats the message according to the format specifier
// and logs it at provided level. Any fields extracted from ctx by the Context
// option are added to the entry.
func (s *SugaredLogger) LogfCtx(ctx context.Context, lvl zapcore.Level, template string, args ...interface{}) {
	s.log(lvl, template, args, s.ctxFields(ctx))
}

// Debugf formats the message according to the format specifier
// and logs it at [DebugLevel].
func (s *SugaredLogger) Debugf(template string, args ...interface{}) {
	s.log(DebugLevel, template, args, nil)
}

// DebugfCtx formats the message according to the format specifier
// and logs it at [DebugLevel]. Any fields extracted from ctx by the Context
// option are added to the entry.
func (s *SugaredLogger) DebugfCtx(ctx context.Context, template string, args ...interface{}) {
	s.log(DebugLevel, template, args, s.ctxFields(ctx))
}

// Infof formats the message according to the format specifier
// and logs it at [InfoLevel].
func (s *SugaredLogger) Infof(template string, args ...interface{}) {
	s.log(InfoLevel, template, args, nil)
}

// InfofCtx formats the message according to the format specifier
// and logs it at [InfoLevel]. Any fields extracted from ctx by the Context
// option are added to the entry.
func (s *SugaredLogger) InfofCtx(ctx context.Context, template string, args ...interface{}) {
	s.log(InfoLevel, template, args, s.ctxFields(ctx))
}

// Warnf formats the message according to the format specifier
// and logs it at [WarnLevel].
func (s *SugaredLogger) Warnf(template string, args ...interface{}) {
	s.log(WarnLevel, template, args, nil)
}

// WarnfCtx formats the message according to the format specifier
// and logs it at [WarnLevel]. Any fields extracted from ctx by the Context
// option are added to the entry.
func (s *SugaredLogger) WarnfCtx(ctx context.Context, template string, args ...interface{}) {
	s.log(WarnLevel, template, args, s.ctxFields(ctx))
}

// Errorf formats the message according to the format specifier
// and logs it at [ErrorLevel].
func (s *SugaredLogger) Errorf(template string, args ...interface{}) {
	s.log(ErrorLevel, template, args, nil)
}

// ErrorfCtx formats the message according to the format specifier
// and logs it at [ErrorLevel]. Any fields extracted from ctx by the Context
// option are added to the entry.
func (s *SugaredLogger) ErrorfCtx(ctx context.Context, template string, args ...interface{}) {
	s.log(ErrorLevel, template, args, s.ctxFields(ctx))
}

// DPanicf formats the message according to the format specifier
// and logs it at [DPanicLevel].
// In development, the logger then panics. (See [DPanicLevel] for details.)
func (s *SugaredLogger) DPanicf(template string, args ...interface{}) {
	s.log(DPanicLevel, template, args, nil)
}

// DPanicfCtx formats the message according to the format specifier
// and logs it at [DPanicLevel]. Any fields extracted from ctx by the Context
// option are added to the entry.
// In development, the logger then panics. (See [DPanicLevel] for details.)
func (s *SugaredLogger) DPanicfCtx(ctx context.Context, template string, args ...interface{}) {
	s.log(DPanicLevel, template, args, s.ctxFields(ctx))
}

// Panicf formats the message according to the format specifier
// and panics.
func (s *SugaredLogger) Panicf(template string, args ...interface{}) {
	s.log(PanicLevel, template, args, nil)
}

// PanicfCtx formats the message according to the format specifier
// and panics. Any fields extracted from ctx by the Context option are added to
// the entry.
func (s *SugaredLogger) PanicfCtx(ctx context.Context, template string, args ...interface{}) {
	s.log(PanicLevel, template, args, s.ctxFields(ctx))
}

// Fatalf formats the message according to the format specifier
// and calls os.Exit.
func (s *SugaredLogger) Fatalf(template string, args ...interface{}) {
	s.log(FatalLevel, template, args, nil)
}

// FatalfCtx formats the message according to the format specifier
// and calls os.Exit. Any fields extracted from ctx by the Context option are
// added to the entry.
func (s *SugaredLogger) FatalfCtx(ctx context.Context, template string, args ...interface{}) {
	s.log(FatalLevel, template, args, s.ctxFields(ctx))
}

// Logw logs a message with some additional context. The variadic key-value
// pairs are treated as they are in With.
func (s *SugaredLogger) Logw(lvl zapcore.Level, msg string, keysAndValues ...interface{}) {
	s.log(lvl, msg, nil, keysAndValues)
}

// LogwCtx logs a message with some additional context. The variadic key-value
// pairs are treated as they are in With. Any fields extracted from ctx by the
// Context option are prepended to the key-value pairs.
func (s *SugaredLogger) LogwCtx(ctx context.Context, lvl zapcore.Level, msg string, keysAndValues ...interface{}) {
	s.log(lvl, msg, nil, append(s.ctxFields(ctx), keysAndValues...))
}

// Debugw logs a message with some additional context. The variadic key-value
// pairs are treated as they are in With.
//
// When debug-level logging is disabled, this is much faster than
//
//	s.With(keysAndValues).Debug(msg)
func (s *SugaredLogger) Debugw(msg string, keysAndValues ...interface{}) {
	s.log(DebugLevel, msg, nil, keysAndValues)
}

// DebugwCtx logs a message with some additional context. The variadic key-value
// pairs are treated as they are in With. Any fields extracted from ctx by the
// Context option are prepended to the key-value pairs.
func (s *SugaredLogger) DebugwCtx(ctx context.Context, msg string, keysAndValues ...interface{}) {
	s.log(DebugLevel, msg, nil, append(s.ctxFields(ctx), keysAndValues...))
}

// Infow logs a message with some additional context. The variadic key-value
// pairs are treated as they are in With.
func (s *SugaredLogger) Infow(msg string, keysAndValues ...interface{}) {
	s.log(InfoLevel, msg, nil, keysAndValues)
}

// InfowCtx logs a message with some additional context. The variadic key-value
// pairs are treated as they are in With. Any fields extracted from ctx by the
// Context option are prepended to the key-value pairs.
func (s *SugaredLogger) InfowCtx(ctx context.Context, msg string, keysAndValues ...interface{}) {
	s.log(InfoLevel, msg, nil, append(s.ctxFields(ctx), keysAndValues...))
}

// Warnw logs a message with some additional context. The variadic key-value
// pairs are treated as they are in With.
func (s *SugaredLogger) Warnw(msg string, keysAndValues ...interface{}) {
	s.log(WarnLevel, msg, nil, keysAndValues)
}

// WarnwCtx logs a message with some additional context. The variadic key-value
// pairs are treated as they are in With. Any fields extracted from ctx by the
// Context option are prepended to the key-value pairs.
func (s *SugaredLogger) WarnwCtx(ctx context.Context, msg string, keysAndValues ...interface{}) {
	s.log(WarnLevel, msg, nil, append(s.ctxFields(ctx), keysAndValues...))
}

// Errorw logs a message with some additional context. The variadic key-value
// pairs are treated as they are in With.
func (s *SugaredLogger) Errorw(msg string, keysAndValues ...interface{}) {
	s.log(ErrorLevel, msg, nil, keysAndValues)
}

// ErrorwCtx logs a message with some additional context. The variadic key-value
// pairs are treated as they are in With. Any fields extracted from ctx by the
// Context option are prepended to the key-value pairs.
func (s *SugaredLogger) ErrorwCtx(ctx context.Context, msg string, keysAndValues ...interface{}) {
	s.log(ErrorLevel, msg, nil, append(s.ctxFields(ctx), keysAndValues...))
}

// DPanicw logs a message with some additional context. In development, the
// logger then panics. (See DPanicLevel for details.) The variadic key-value
// pairs are treated as they are in With.
func (s *SugaredLogger) DPanicw(msg string, keysAndValues ...interface{}) {
	s.log(DPanicLevel, msg, nil, keysAndValues)
}

// DPanicwCtx logs a message with some additional context. In development, the
// logger then panics. (See DPanicLevel for details.) The variadic key-value
// pairs are treated as they are in With. Any fields extracted from ctx by the
// Context option are prepended to the key-value pairs.
func (s *SugaredLogger) DPanicwCtx(ctx context.Context, msg string, keysAndValues ...interface{}) {
	s.log(DPanicLevel, msg, nil, append(s.ctxFields(ctx), keysAndValues...))
}

// Panicw logs a message with some additional context, then panics. The
// variadic key-value pairs are treated as they are in With.
func (s *SugaredLogger) Panicw(msg string, keysAndValues ...interface{}) {
	s.log(PanicLevel, msg, nil, keysAndValues)
}

// PanicwCtx logs a message with some additional context, then panics. The
// variadic key-value pairs are treated as they are in With. Any fields
// extracted from ctx by the Context option are prepended to the key-value
// pairs.
func (s *SugaredLogger) PanicwCtx(ctx context.Context, msg string, keysAndValues ...interface{}) {
	s.log(PanicLevel, msg, nil, append(s.ctxFields(ctx), keysAndValues...))
}

// Fatalw logs a message with some additional context, then calls os.Exit. The
// variadic key-value pairs are treated as they are in With.
func (s *SugaredLogger) Fatalw(msg string, keysAndValues ...interface{}) {
	s.log(FatalLevel, msg, nil, keysAndValues)
}

// FatalwCtx logs a message with some additional context, then calls os.Exit.
// The variadic key-value pairs are treated as they are in With. Any fields
// extracted from ctx by the Context option are prepended to the key-value
// pairs.
func (s *SugaredLogger) FatalwCtx(ctx context.Context, msg string, keysAndValues ...interface{}) {
	s.log(FatalLevel, msg, nil, append(s.ctxFields(ctx), keysAndValues...))
}

// Logln logs a message at provided level.
// Spaces are always added between arguments.
func (s *SugaredLogger) Logln(lvl zapcore.Level, args ...interface{}) {
	s.logln(lvl, args, nil)
}

// LoglnCtx logs a message at provided level. Any fields extracted from ctx by
// the Context option are added to the entry.
// Spaces are always added between arguments.
func (s *SugaredLogger) LoglnCtx(ctx context.Context, lvl zapcore.Level, args ...interface{}) {
	s.logln(lvl, args, s.ctxFields(ctx))
}

// Debugln logs a message at [DebugLevel].
// Spaces are always added between arguments.
func (s *SugaredLogger) Debugln(args ...interface{}) {
	s.logln(DebugLevel, args, nil)
}

// DebuglnCtx logs a message at [DebugLevel]. Any fields extracted from ctx by
// the Context option are added to the entry.
// Spaces are always added between arguments.
func (s *SugaredLogger) DebuglnCtx(ctx context.Context, args ...interface{}) {
	s.logln(DebugLevel, args, s.ctxFields(ctx))
}

// Infoln logs a message at [InfoLevel].
// Spaces are always added between arguments.
func (s *SugaredLogger) Infoln(args ...interface{}) {
	s.logln(InfoLevel, args, nil)
}

// InfolnCtx logs a message at [InfoLevel]. Any fields extracted from ctx by
// the Context option are added to the entry.
// Spaces are always added between arguments.
func (s *SugaredLogger) InfolnCtx(ctx context.Context, args ...interface{}) {
	s.logln(InfoLevel, args, s.ctxFields(ctx))
}

// Warnln logs a message at [WarnLevel].
// Spaces are always added between arguments.
func (s *SugaredLogger) Warnln(args ...interface{}) {
	s.logln(WarnLevel, args, nil)
}

// WarnlnCtx logs a message at [WarnLevel]. Any fields extracted from ctx by
// the Context option are added to the entry.
// Spaces are always added between arguments.
func (s *SugaredLogger) WarnlnCtx(ctx context.Context, args ...interface{}) {
	s.logln(WarnLevel, args, s.ctxFields(ctx))
}

// Errorln logs a message at [ErrorLevel].
// Spaces are always added between arguments.
func (s *SugaredLogger) Errorln(args ...interface{}) {
	s.logln(ErrorLevel, args, nil)
}

// ErrorlnCtx logs a message at [ErrorLevel]. Any fields extracted from ctx by
// the Context option are added to the entry.
// Spaces are always added between arguments.
func (s *SugaredLogger) ErrorlnCtx(ctx context.Context, args ...interface{}) {
	s.logln(ErrorLevel, args, s.ctxFields(ctx))
}

// DPanicln logs a message at [DPanicLevel].
// In development, the logger then panics. (See [DPanicLevel] for details.)
// Spaces are always added between arguments.
func (s *SugaredLogger) DPanicln(args ...interface{}) {
	s.logln(DPanicLevel, args, nil)
}

// DPaniclnCtx logs a message at [DPanicLevel]. Any fields extracted from ctx by
// the Context option are added to the entry.
// In development, the logger then panics. (See [DPanicLevel] for details.)
// Spaces are always added between arguments.
func (s *SugaredLogger) DPaniclnCtx(ctx context.Context, args ...interface{}) {
	s.logln(DPanicLevel, args, s.ctxFields(ctx))
}

// Panicln logs a message at [PanicLevel] and panics.
// Spaces are always added between arguments.
func (s *SugaredLogger) Panicln(args ...interface{}) {
	s.logln(PanicLevel, args, nil)
}

// PaniclnCtx logs a message at [PanicLevel] and panics. Any fields extracted
// from ctx by the Context option are added to the entry.
// Spaces are always added between arguments.
func (s *SugaredLogger) PaniclnCtx(ctx context.Context, args ...interface{}) {
	s.logln(PanicLevel, args, s.ctxFields(ctx))
}

// Fatalln logs a message at [FatalLevel] and calls os.Exit.
// Spaces are always added between arguments.
func (s *SugaredLogger) Fatalln(args ...interface{}) {
	s.logln(FatalLevel, args, nil)
}

// FatallnCtx logs a message at [FatalLevel] and calls os.Exit. Any fields
// extracted from ctx by the Context option are added to the entry.
// Spaces are always added between arguments.
func (s *SugaredLogger) FatallnCtx(ctx context.Context, args ...interface{}) {
	s.logln(FatalLevel, args, s.ctxFields(ctx))
}

// Sync flushes any buffered log entries.
func (s *SugaredLogger) Sync() error {
	return s.base.Sync()
}

// ctxFields returns the fields extracted from ctx by the configured Context
// option, as a slice of loosely-typed arguments suitable for the SugaredLogger
// log path. It returns nil when no Context option is set or ctx is nil.
func (s *SugaredLogger) ctxFields(ctx context.Context) []interface{} {
	if ctx == nil || s.base.contextFunc == nil {
		return nil
	}
	fields := s.base.contextFunc(ctx)
	args := make([]interface{}, 0, len(fields))
	for _, f := range fields {
		args = append(args, f)
	}
	return args
}

// log message with Sprint, Sprintf, or neither.
func (s *SugaredLogger) log(lvl zapcore.Level, template string, fmtArgs []interface{}, context []interface{}) {
	// If logging at this level is completely disabled, skip the overhead of
	// string formatting.
	if lvl < DPanicLevel && !s.base.Core().Enabled(lvl) {
		return
	}

	msg := getMessage(template, fmtArgs)
	if ce := s.base.Check(lvl, msg); ce != nil {
		ce.Write(s.sweetenFields(context)...)
	}
}

// logln message with Sprintln
func (s *SugaredLogger) logln(lvl zapcore.Level, fmtArgs []interface{}, context []interface{}) {
	if lvl < DPanicLevel && !s.base.Core().Enabled(lvl) {
		return
	}

	msg := getMessageln(fmtArgs)
	if ce := s.base.Check(lvl, msg); ce != nil {
		ce.Write(s.sweetenFields(context)...)
	}
}

// getMessage format with Sprint, Sprintf, or neither.
func getMessage(template string, fmtArgs []interface{}) string {
	if len(fmtArgs) == 0 {
		return template
	}

	if template != "" {
		return fmt.Sprintf(template, fmtArgs...)
	}

	if len(fmtArgs) == 1 {
		if str, ok := fmtArgs[0].(string); ok {
			return str
		}
	}
	return fmt.Sprint(fmtArgs...)
}

// getMessageln format with Sprintln.
func getMessageln(fmtArgs []interface{}) string {
	msg := fmt.Sprintln(fmtArgs...)
	return msg[:len(msg)-1]
}

func (s *SugaredLogger) sweetenFields(args []interface{}) []Field {
	if len(args) == 0 {
		return nil
	}

	var (
		// Allocate enough space for the worst case; if users pass only structured
		// fields, we shouldn't penalize them with extra allocations.
		fields    = make([]Field, 0, len(args))
		invalid   invalidPairs
		seenError bool
	)

	for i := 0; i < len(args); {
		// This is a strongly-typed field. Consume it and move on.
		if f, ok := args[i].(Field); ok {
			fields = append(fields, f)
			i++
			continue
		}

		// If it is an error, consume it and move on.
		if err, ok := args[i].(error); ok {
			if !seenError {
				seenError = true
				fields = append(fields, Error(err))
			} else {
				s.base.Error(_multipleErrMsg, Error(err))
			}
			i++
			continue
		}

		// Make sure this element isn't a dangling key.
		if i == len(args)-1 {
			s.base.Error(_oddNumberErrMsg, Any("ignored", args[i]))
			break
		}

		// Consume this value and the next, treating them as a key-value pair. If the
		// key isn't a string, add this pair to the slice of invalid pairs.
		key, val := args[i], args[i+1]
		if keyStr, ok := key.(string); !ok {
			// Subsequent errors are likely, so allocate once up front.
			if cap(invalid) == 0 {
				invalid = make(invalidPairs, 0, len(args)/2)
			}
			invalid = append(invalid, invalidPair{i, key, val})
		} else {
			fields = append(fields, Any(keyStr, val))
		}
		i += 2
	}

	// If we encountered any invalid key-value pairs, log an error.
	if len(invalid) > 0 {
		s.base.Error(_nonStringKeyErrMsg, Array("invalid", invalid))
	}
	return fields
}

type invalidPair struct {
	position   int
	key, value interface{}
}

func (p invalidPair) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddInt64("position", int64(p.position))
	Any("key", p.key).AddTo(enc)
	Any("value", p.value).AddTo(enc)
	return nil
}

type invalidPairs []invalidPair

func (ps invalidPairs) MarshalLogArray(enc zapcore.ArrayEncoder) error {
	var err error
	for i := range ps {
		err = multierr.Append(err, enc.AppendObject(ps[i]))
	}
	return err
}
