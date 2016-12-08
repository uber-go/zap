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
	"time"
)

// For tests.
var _exit = os.Exit

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

	// Facility returns the destination that log entrys are written to.
	Facility() Facility
}

type logger struct {
	fac Facility

	development bool
	hooks       []hook
	errorOutput WriteSyncer
}

// New returns a new logger with sensible defaults: logging at InfoLevel,
// development mode off, errors writtten to standard error, and logs JSON
// encoded to standard output.
func New(fac Facility, options ...Option) Logger {
	if fac == nil {
		fac = WriterFacility(NewJSONEncoder(), nil, InfoLevel)
	}
	log := &logger{
		fac:         fac,
		errorOutput: newLockedWriteSyncer(os.Stderr),
	}
	for _, opt := range options {
		opt.apply(log)
	}
	return log
}

func (log *logger) With(fields ...Field) Logger {
	return &logger{
		fac:         log.fac.With(fields...),
		development: log.development,
		hooks:       log.hooks,
		errorOutput: log.errorOutput,
	}
}

func (log *logger) Check(lvl Level, msg string) *CheckedEntry {
	ent := Entry{
		Time:    time.Now().UTC(),
		Level:   lvl,
		Message: msg,
	}
	ce := log.fac.Check(ent, nil)
	switch ent.Level {
	case PanicLevel:
		// Panic should always cause a panic, even if not written.
		return ce.Should(ent, WriteThenPanic)
	case FatalLevel:
		// Fatal should always cause an exit.
		return ce.Should(ent, WriteThenFatal)
	case DPanicLevel:
		// DPanic should always cause a panic in development.
		if log.development {
			return ce.Should(ent, WriteThenPanic)
		}
		fallthrough
	default:
		return ce
	}
}

func (log *logger) Debug(msg string, fields ...Field) {
	log.Log(DebugLevel, msg, fields...)
}

func (log *logger) Info(msg string, fields ...Field) {
	log.Log(InfoLevel, msg, fields...)
}

func (log *logger) Warn(msg string, fields ...Field) {
	log.Log(WarnLevel, msg, fields...)
}

func (log *logger) Error(msg string, fields ...Field) {
	log.Log(ErrorLevel, msg, fields...)
}

func (log *logger) DPanic(msg string, fields ...Field) {
	log.Log(DPanicLevel, msg, fields...)
	if log.development {
		panic(msg)
	}
}

func (log *logger) Panic(msg string, fields ...Field) {
	log.Log(PanicLevel, msg, fields...)
	panic(msg)
}

func (log *logger) Fatal(msg string, fields ...Field) {
	log.Log(FatalLevel, msg, fields...)
	_exit(1)
}

func (log *logger) Log(lvl Level, msg string, fields ...Field) {
	ent := Entry{
		Time:    time.Now().UTC(),
		Level:   lvl,
		Message: msg,
	}
	for _, hook := range log.hooks {
		if hookedEnt, err := hook(ent); err != nil {
			log.InternalError("hook", err)
		} else {
			ent = hookedEnt
		}
	}
	if err := log.fac.Log(ent, fields...); err != nil {
		log.InternalError("facility", err)
	}
}

// InternalError prints an internal error message to the configured
// ErrorOutput. This method should only be used to report internal logger
// problems and should not be used to report user-caused problems.
func (log *logger) InternalError(cause string, err error) {
	fmt.Fprintf(log.errorOutput, "%v %s error: %v\n", time.Now().UTC(), cause, err)
	log.errorOutput.Sync()
}

// Facility returns the destination that logs entries are written to.
func (log *logger) Facility() Facility {
	return log.fac
}
