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

	"github.com/uber-go/atomic"
)

// For tests.
var _exit = os.Exit

// A Logger enables leveled, structured logging. All methods are safe for
// concurrent use.
type Logger interface {
	// Check the minimum enabled log level.
	Level() Level
	// Change the level of this logger, as well as all its ancestors and
	// descendants. This makes it easy to change the log level at runtime
	// without restarting your application.
	SetLevel(Level)
	// Create a child logger, and optionally add some context to that logger.
	With(...Field) Logger
	// StubTime stops the logger from including the current time in each
	// message. Instead, it always reports the time as Unix epoch 0. (This is
	// useful in tests and examples.)
	//
	// TODO: remove this kludge in favor of a more comprehensive message-formatting
	// option.
	StubTime()

	// Check returns a CheckedMessage if logging a message at the specified level
	// is enabled. It's a completely optional optimization; in high-performance
	// applications, Check can help avoid allocating a slice to hold fields.
	//
	// See CheckedMessage for an example.
	Check(Level, string) *CheckedMessage

	// Log a message at the given level. Messages include any context that's
	// accumulated on the logger, as well as any fields added at the log site.
	Log(Level, string, ...Field)
	Debug(string, ...Field)
	Info(string, ...Field)
	Warn(string, ...Field)
	Error(string, ...Field)
	Panic(string, ...Field)
	Fatal(string, ...Field)
	// If the logger is in development mode (via the Development option), DFatal
	// logs at the Fatal level. Otherwise, it logs at the Error level.
	DFatal(string, ...Field)
}

type jsonLogger struct {
	level       *atomic.Int32
	enc         encoder
	hooks       []hook
	errW        WriteSyncer
	w           WriteSyncer
	alwaysEpoch bool
	development bool
}

// NewJSON returns a logger that formats its output as JSON. Zap uses a
// customized JSON encoder to avoid reflection and minimize allocations.
//
// By default, the logger will write Info logs or higher to standard
// out. Any errors during logging will be written to standard error.
//
// Options can change the log level, the output location, or the initial
// fields that should be added as context.
func NewJSON(options ...Option) Logger {
	defaultLevel := atomic.NewInt32(int32(InfoLevel))
	jl := &jsonLogger{
		enc:   newJSONEncoder(),
		level: defaultLevel,
		errW:  os.Stderr,
		w:     os.Stdout,
	}

	for _, opt := range options {
		opt.apply(jl)
	}

	return jl
}

func (jl *jsonLogger) Level() Level {
	return Level(jl.level.Load())
}

func (jl *jsonLogger) SetLevel(lvl Level) {
	jl.level.Store(int32(lvl))
}

func (jl *jsonLogger) With(fields ...Field) Logger {
	clone := &jsonLogger{
		level:       jl.level,
		enc:         jl.enc.Clone(),
		hooks:       jl.hooks,
		w:           jl.w,
		errW:        jl.errW,
		alwaysEpoch: jl.alwaysEpoch,
		development: jl.development,
	}
	clone.enc.AddFields(fields)
	return clone
}

func (jl *jsonLogger) StubTime() {
	jl.alwaysEpoch = true
}

func (jl *jsonLogger) Check(lvl Level, msg string) *CheckedMessage {
	if !(lvl >= jl.Level()) {
		return nil
	}
	return NewCheckedMessage(jl, lvl, msg)
}

func (jl *jsonLogger) Log(lvl Level, msg string, fields ...Field) {
	switch lvl {
	case PanicLevel:
		jl.Panic(msg, fields...)
	case FatalLevel:
		jl.Fatal(msg, fields...)
	default:
		jl.log(lvl, msg, fields)
	}
}

func (jl *jsonLogger) Debug(msg string, fields ...Field) {
	jl.log(DebugLevel, msg, fields)
}

func (jl *jsonLogger) Info(msg string, fields ...Field) {
	jl.log(InfoLevel, msg, fields)
}

func (jl *jsonLogger) Warn(msg string, fields ...Field) {
	jl.log(WarnLevel, msg, fields)
}

func (jl *jsonLogger) Error(msg string, fields ...Field) {
	jl.log(ErrorLevel, msg, fields)
}

func (jl *jsonLogger) Panic(msg string, fields ...Field) {
	jl.log(PanicLevel, msg, fields)
	panic(msg)
}

func (jl *jsonLogger) Fatal(msg string, fields ...Field) {
	jl.log(FatalLevel, msg, fields)
	_exit(1)
}

func (jl *jsonLogger) DFatal(msg string, fields ...Field) {
	if jl.development {
		jl.Fatal(msg, fields...)
	}
	jl.Error(msg, fields...)
}

func (jl *jsonLogger) log(lvl Level, msg string, fields []Field) {
	if !(lvl >= jl.Level()) {
		return
	}

	temp := jl.enc.Clone()
	temp.AddFields(fields)

	for _, hook := range jl.hooks {
		newMsg, err := hook(lvl, msg, temp)
		if err != nil {
			jl.internalError(err.Error())
		}
		msg = newMsg
	}

	now := time.Now().UTC()
	if jl.alwaysEpoch {
		now = time.Unix(0, 0)
	}
	if err := temp.WriteMessage(jl.w, lvl.String(), msg, now); err != nil {
		jl.internalError(err.Error())
	}
	temp.Free()

	if lvl > ErrorLevel {
		// Sync on Panic and Fatal, since they may crash the program.
		jl.w.Sync()
	}
}

func (jl *jsonLogger) internalError(msg string) {
	fmt.Fprintln(jl.errW, msg)
	jl.errW.Sync()
}
