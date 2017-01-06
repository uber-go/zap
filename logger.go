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
)

var _exit = os.Exit // for tests

// A Logger enables leveled, structured logging. All methods are safe for
// concurrent use.
type Logger interface {
	// Create a child logger, and optionally add some context to that logger.
	With(...Field) Logger

	// Check returns a CheckedEntry if logging a message at the specified level
	// is enabled. It's a completely optional optimization; in high-performance
	// applications, Check can help avoid allocating a slice to hold fields.
	//
	// See CheckedEntry for an example.
	Check(Level, string) *CheckedEntry

	// Log a message at the given level. Messages include any context that's
	// accumulated on the logger, as well as any fields added at the log site.
	//
	// Calling Panic should panic() and calling Fatal should terminate the
	// process, but calling Log(PanicLevel, ...) or Log(FatalLevel, ...) should
	// not. It may not be possible for compatibility wrappers to comply with
	// this last part (e.g. the bark wrapper).
	Debug(string, ...Field)
	Info(string, ...Field)
	Warn(string, ...Field)
	Error(string, ...Field)
	DPanic(string, ...Field)
	Panic(string, ...Field)
	Fatal(string, ...Field)

	// Facility returns the destination that log entries are written to.
	Facility() Facility
}

type logger struct {
	fac Facility

	development bool
	errorOutput WriteSyncer

	// TODO: consider using a LevelEnabler instead
	addCaller bool
	addStack  Level

	callerSkip int
}

// New returns a new logger with sensible defaults: logging at InfoLevel,
// development mode off, errors written to standard error, and logs JSON
// encoded to standard output.
func New(fac Facility, options ...Option) Logger {
	if fac == nil {
		fac = WriterFacility(NewJSONEncoder(), os.Stdout, InfoLevel)
	}
	log := &logger{
		fac:         fac,
		errorOutput: newLockedWriteSyncer(os.Stderr),
		addStack:    maxLevel, // TODO: better an `always false` level enabler
	}
	for _, opt := range options {
		opt.apply(log)
	}
	return log
}

func (log *logger) With(fields ...Field) Logger {
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

func (log *logger) Check(lvl Level, msg string) *CheckedEntry {
	return log.check(lvl, msg)
}

func (log *logger) check(lvl Level, msg string) *CheckedEntry {
	// Create basic checked entry thru the facility; this will be non-nil if
	// the log message will actually be written somewhere.
	ent := Entry{
		Time:    time.Now().UTC(),
		Level:   lvl,
		Message: msg,
	}
	ce := log.fac.Check(ent, nil)
	willWrite := ce != nil

	// If terminal behavior is required, setup so that it happens after the
	// checked entry is written and create a checked entry if it's still nil.
	switch ent.Level {
	case PanicLevel:
		ce = ce.Should(ent, WriteThenPanic)
	case FatalLevel:
		ce = ce.Should(ent, WriteThenFatal)
	case DPanicLevel:
		if log.development {
			ce = ce.Should(ent, WriteThenPanic)
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
		ce.Entry.Caller = MakeEntryCaller(runtime.Caller(log.callerSkip + 2))
		if !ce.Entry.Caller.Defined {
			fmt.Fprintf(log.errorOutput, "%v addCaller error: failed to get caller\n", time.Now().UTC())
			log.errorOutput.Sync()
		}
	}

	if ce.Entry.Level >= log.addStack {
		ce.Entry.Stack = Stack().str
		// TODO: maybe just inline Stack around takeStacktrace
	}

	return ce
}

func (log *logger) Debug(msg string, fields ...Field) {
	if ce := log.check(DebugLevel, msg); ce != nil {
		ce.Write(fields...)
	}
}

func (log *logger) Info(msg string, fields ...Field) {
	if ce := log.check(InfoLevel, msg); ce != nil {
		ce.Write(fields...)
	}
}

func (log *logger) Warn(msg string, fields ...Field) {
	if ce := log.check(WarnLevel, msg); ce != nil {
		ce.Write(fields...)
	}
}

func (log *logger) Error(msg string, fields ...Field) {
	if ce := log.check(ErrorLevel, msg); ce != nil {
		ce.Write(fields...)
	}
}

func (log *logger) DPanic(msg string, fields ...Field) {
	if ce := log.check(DPanicLevel, msg); ce != nil {
		ce.Write(fields...)
	}
}

func (log *logger) Panic(msg string, fields ...Field) {
	if ce := log.check(PanicLevel, msg); ce != nil {
		ce.Write(fields...)
	}
}

func (log *logger) Fatal(msg string, fields ...Field) {
	if ce := log.check(FatalLevel, msg); ce != nil {
		ce.Write(fields...)
	}
}

func (log *logger) Log(lvl Level, msg string, fields ...Field) {
	if ce := log.check(lvl, msg); ce != nil {
		ce.Write(fields...)
	}
}

// Facility returns the destination that logs entries are written to.
func (log *logger) Facility() Facility {
	return log.fac
}
