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
	"time"

	"go.uber.org/zap/zapcore"
)

func defaultEncoderConfig() zapcore.JSONConfig {
	msgF := func(msg string) zapcore.Field {
		return zapcore.Field{Type: zapcore.StringType, String: msg, Key: "msg"}
	}
	timeF := func(t time.Time) zapcore.Field {
		millis := t.UnixNano() / int64(time.Millisecond)
		return zapcore.Field{Type: zapcore.Int64Type, Integer: millis, Key: "ts"}
	}
	levelF := func(l zapcore.Level) zapcore.Field {
		return zapcore.Field{Type: zapcore.StringType, String: l.String(), Key: "level"}
	}
	return zapcore.JSONConfig{
		MessageFormatter: msgF,
		TimeFormatter:    timeF,
		LevelFormatter:   levelF,
	}
}

// A Logger enables leveled, structured logging. All methods are safe for
// concurrent use.
type Logger interface {
	// Create a child logger, and optionally add some context to that logger.
	With(...zapcore.Field) Logger

	// Check returns a CheckedEntry if logging a message at the specified level
	// is enabled. It's a completely optional optimization; in high-performance
	// applications, Check can help avoid allocating a slice to hold fields.
	Check(zapcore.Level, string) *zapcore.CheckedEntry

	// Log a message at the given level. Messages include any context that's
	// accumulated on the logger, as well as any fields added at the log site.
	//
	// Calling Panic should panic and calling Fatal should terminate the process,
	// even if logging at those levels is disabled.
	Debug(string, ...zapcore.Field)
	Info(string, ...zapcore.Field)
	Warn(string, ...zapcore.Field)
	Error(string, ...zapcore.Field)
	DPanic(string, ...zapcore.Field)
	Panic(string, ...zapcore.Field)
	Fatal(string, ...zapcore.Field)

	// Facility returns the destination that log entries are written to.
	Facility() zapcore.Facility
}

type logger struct {
	fac zapcore.Facility

	development bool
	errorOutput zapcore.WriteSyncer

	addCaller bool
	addStack  zapcore.LevelEnabler

	callerSkip int
}

// New returns a new logger with sensible defaults: logging at InfoLevel,
// development mode off, errors written to standard error, and logs JSON
// encoded to standard output.
func New(fac zapcore.Facility, options ...Option) Logger {
	if fac == nil {
		fac = zapcore.WriterFacility(
			zapcore.NewJSONEncoder(defaultEncoderConfig()),
			os.Stdout,
			InfoLevel,
		)
	}
	log := &logger{
		fac:         fac,
		errorOutput: zapcore.Lock(os.Stderr),
		addStack:    LevelEnablerFunc(func(_ zapcore.Level) bool { return false }),
	}
	for _, opt := range options {
		opt.apply(log)
	}
	return log
}

func (log *logger) With(fields ...zapcore.Field) Logger {
	if len(fields) == 0 {
		return log
	}
	return &logger{
		fac:         log.fac.With(fields),
		development: log.development,
		errorOutput: log.errorOutput,
		addCaller:   log.addCaller,
		addStack:    log.addStack,
		callerSkip:  log.callerSkip,
	}
}

func (log *logger) Check(lvl zapcore.Level, msg string) *zapcore.CheckedEntry {
	return log.check(lvl, msg)
}

func (log *logger) check(lvl zapcore.Level, msg string) *zapcore.CheckedEntry {
	// check must always be called directly by a method in the Logger interface
	// (e.g., Check, Info, Fatal).
	const callerSkipOffset = 2

	// Create basic checked entry thru the facility; this will be non-nil if
	// the log message will actually be written somewhere.
	ent := zapcore.Entry{
		Time:    time.Now().UTC(),
		Level:   lvl,
		Message: msg,
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

func (log *logger) Debug(msg string, fields ...zapcore.Field) {
	if ce := log.check(DebugLevel, msg); ce != nil {
		ce.Write(fields...)
	}
}

func (log *logger) Info(msg string, fields ...zapcore.Field) {
	if ce := log.check(InfoLevel, msg); ce != nil {
		ce.Write(fields...)
	}
}

func (log *logger) Warn(msg string, fields ...zapcore.Field) {
	if ce := log.check(WarnLevel, msg); ce != nil {
		ce.Write(fields...)
	}
}

func (log *logger) Error(msg string, fields ...zapcore.Field) {
	if ce := log.check(ErrorLevel, msg); ce != nil {
		ce.Write(fields...)
	}
}

func (log *logger) DPanic(msg string, fields ...zapcore.Field) {
	if ce := log.check(DPanicLevel, msg); ce != nil {
		ce.Write(fields...)
	}
}

func (log *logger) Panic(msg string, fields ...zapcore.Field) {
	if ce := log.check(PanicLevel, msg); ce != nil {
		ce.Write(fields...)
	}
}

func (log *logger) Fatal(msg string, fields ...zapcore.Field) {
	if ce := log.check(FatalLevel, msg); ce != nil {
		ce.Write(fields...)
	}
}

// Facility returns the destination that logs entries are written to.
func (log *logger) Facility() zapcore.Facility {
	return log.fac
}
