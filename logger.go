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
	"os"
	"runtime"
	"strings"
	"time"

	"go.uber.org/zap/zapcore"
)

func defaultEncoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		MessageKey:     "msg",
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "name",
		CallerKey:      "caller",
		StacktraceKey:  "stacktrace",
		EncodeTime:     func(t time.Time, enc zapcore.ArrayEncoder) { enc.AppendInt64(t.UnixNano() / int64(time.Millisecond)) },
		EncodeDuration: func(d time.Duration, enc zapcore.ArrayEncoder) { enc.AppendInt64(int64(d)) },
		EncodeLevel:    func(l zapcore.Level, enc zapcore.ArrayEncoder) { enc.AppendString(l.String()) },
	}
}

// A Logger provides fast, leveled, structured logging. All methods are safe for
// concurrent use.
//
// If you'd prefer a slower but less verbose logger, see the SugaredLogger.
type Logger struct {
	fac zapcore.Facility

	development bool
	name        string
	errorOutput zapcore.WriteSyncer

	addCaller bool
	addStack  zapcore.LevelEnabler

	callerSkip int
}

// New returns a new logger with sensible defaults: logging at InfoLevel,
// development mode off, errors written to standard error, and logs JSON
// encoded to standard output.
func New(fac zapcore.Facility, options ...Option) *Logger {
	if fac == nil {
		fac = zapcore.WriterFacility(
			zapcore.NewJSONEncoder(defaultEncoderConfig()),
			os.Stdout,
			InfoLevel,
		)
	}
	log := &Logger{
		fac:         fac,
		errorOutput: zapcore.Lock(os.Stderr),
		addStack:    LevelEnablerFunc(func(_ zapcore.Level) bool { return false }),
	}
	for _, opt := range options {
		opt.apply(log)
	}
	return log
}

// Sugar converts a Logger to a SugaredLogger.
func (log *Logger) Sugar() *SugaredLogger {
	core := log.clone()
	core.callerSkip += 2
	return &SugaredLogger{core}
}

// Named adds a new path segment to the logger's name. Segments are joined by
// periods.
func (log *Logger) Named(s string) *Logger {
	if s == "" {
		return log
	}
	l := log.clone()
	if log.name == "" {
		l.name = s
	} else {
		l.name = strings.Join([]string{l.name, s}, ".")
	}
	return l
}

// With creates a child logger and adds structured context to it. Fields added
// to the child don't affect the parent, and vice versa.
func (log *Logger) With(fields ...zapcore.Field) *Logger {
	if len(fields) == 0 {
		return log
	}
	l := log.clone()
	l.fac = l.fac.With(fields)
	return l
}

// Check returns a CheckedEntry if logging a message at the specified level
// is enabled. It's a completely optional optimization; in high-performance
// applications, Check can help avoid allocating a slice to hold fields.
func (log *Logger) Check(lvl zapcore.Level, msg string) *zapcore.CheckedEntry {
	return log.check(lvl, msg)
}

// Debug logs a message at DebugLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func (log *Logger) Debug(msg string, fields ...zapcore.Field) {
	if ce := log.check(DebugLevel, msg); ce != nil {
		ce.Write(fields...)
	}
}

// Info logs a message at InfoLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func (log *Logger) Info(msg string, fields ...zapcore.Field) {
	if ce := log.check(InfoLevel, msg); ce != nil {
		ce.Write(fields...)
	}
}

// Warn logs a message at WarnLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func (log *Logger) Warn(msg string, fields ...zapcore.Field) {
	if ce := log.check(WarnLevel, msg); ce != nil {
		ce.Write(fields...)
	}
}

// Error logs a message at ErrorLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func (log *Logger) Error(msg string, fields ...zapcore.Field) {
	if ce := log.check(ErrorLevel, msg); ce != nil {
		ce.Write(fields...)
	}
}

// DPanic logs a message at DPanicLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
//
// If the logger is in development mode, it then panics (DPanic means
// "development panic"). This is useful for catching errors that are
// recoverable, but shouldn't ever happen.
func (log *Logger) DPanic(msg string, fields ...zapcore.Field) {
	if ce := log.check(DPanicLevel, msg); ce != nil {
		ce.Write(fields...)
	}
}

// Panic logs a message at PanicLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
//
// The logger then panics, even if logging at PanicLevel is disabled.
func (log *Logger) Panic(msg string, fields ...zapcore.Field) {
	if ce := log.check(PanicLevel, msg); ce != nil {
		ce.Write(fields...)
	}
}

// Fatal logs a message at FatalLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
//
// The logger then calls os.Exit(1), even if logging at FatalLevel is disabled.
func (log *Logger) Fatal(msg string, fields ...zapcore.Field) {
	if ce := log.check(FatalLevel, msg); ce != nil {
		ce.Write(fields...)
	}
}

// Facility returns the destination that logs entries are written to.
func (log *Logger) Facility() zapcore.Facility {
	return log.fac
}

func (log *Logger) clone() *Logger {
	copy := *log
	return &copy
}

func (log *Logger) check(lvl zapcore.Level, msg string) *zapcore.CheckedEntry {
	// check must always be called directly by a method in the Logger interface
	// (e.g., Check, Info, Fatal).
	const callerSkipOffset = 2

	// Create basic checked entry thru the facility; this will be non-nil if
	// the log message will actually be written somewhere.
	ent := zapcore.Entry{
		LoggerName: log.name,
		Time:       time.Now().UTC(),
		Level:      lvl,
		Message:    msg,
	}
	ce := log.fac.Check(ent, nil)

	willWrite := ce != nil

	// If terminal behavior is required, setup so that it happens after the
	// checked entry is written and create a checked entry if it's still nil.
	switch ent.Level {
	case zapcore.PanicLevel:
		ce = ce.Should(ent, zapcore.WriteThenPanic)
	case zapcore.FatalLevel:
		ce = ce.Should(ent, zapcore.WriteThenFatal)
	case zapcore.DPanicLevel:
		if log.development {
			ce = ce.Should(ent, zapcore.WriteThenPanic)
		}
	}

	// Only do further annotation if we're going to write this message; checked
	// entries that exist only for terminal behavior do not benefit from
	// annotation.
	if !willWrite {
		return ce
	}

	ce.ErrorOutput = log.errorOutput
	if log.addCaller {
		ce.Entry.Caller = zapcore.MakeEntryCaller(runtime.Caller(log.callerSkip + callerSkipOffset))
		if !ce.Entry.Caller.Defined {
			fmt.Fprintf(log.errorOutput, "%v addCaller error: failed to get caller\n", time.Now().UTC())
			log.errorOutput.Sync()
		}
	}

	if log.addStack.Enabled(ce.Entry.Level) {
		ce.Entry.Stack = Stack().String
		// TODO: maybe just inline Stack around takeStacktrace
	}

	return ce
}
