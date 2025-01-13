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

func filterOptions[T Option](in ...Option) []T {
	var opts []T
	for _, opt := range in {
		if o, ok := opt.(T); ok {
			opts = append(opts, o)
		}
	}
	return opts
}

// An Option configures a Logger.
type Option interface {
	apply(*Logger)
}

// wrapEncoderOption wraps a func to make it satisfy the WrapEncoderOption interface.
type wrapEncoderOption func(encoding string, cfg zapcore.EncoderConfig, encoder zapcore.Encoder) (zapcore.Encoder, error)

func (f wrapEncoderOption) wrapEncoder(encoding string, cfg zapcore.EncoderConfig, encoder zapcore.Encoder) (zapcore.Encoder, error) {
	return f(encoding, cfg, encoder)
}

func (f wrapEncoderOption) apply(_ *Logger) {}

// WrapEncoder wraps or replaces the Logger's underlying zapcore.Encoder.
func WrapEncoder(fn func(encoding string, cfg zapcore.EncoderConfig, encoder zapcore.Encoder) (zapcore.Encoder, error)) Option {
	return wrapEncoderOption(fn)
}

// wrapEncoderOption wraps a func to make it satisfy the WrapSinkersOption interface.
type wrapSinkerOption func(paths []string, sink zapcore.WriteSyncer, errPath []string, errSink zapcore.WriteSyncer) (zapcore.WriteSyncer, zapcore.WriteSyncer, error)

func (f wrapSinkerOption) wrapSink(paths []string, sink zapcore.WriteSyncer, errPath []string, errSink zapcore.WriteSyncer) (zapcore.WriteSyncer, zapcore.WriteSyncer, error) {
	return f(paths, sink, errPath, errSink)
}

func (f wrapSinkerOption) apply(_ *Logger) {}

// WrapSinker wraps or replaces the Logger's underlying zapcore.Encoder.
func WrapSinker(fn func(paths []string, sink zapcore.WriteSyncer, errPath []string, errSink zapcore.WriteSyncer) (zapcore.WriteSyncer, zapcore.WriteSyncer, error)) Option {
	return wrapSinkerOption(fn)
}

// optionFunc wraps a func so it satisfies the Option interface.
type optionFunc func(*Logger)

func (f optionFunc) apply(log *Logger) {
	f(log)
}

// WrapCore wraps or replaces the Logger's underlying zapcore.Core.
func WrapCore(f func(zapcore.Core) zapcore.Core) Option {
	return optionFunc(func(log *Logger) {
		log.core = f(log.core)
	})
}

// Hooks registers functions which will be called each time the Logger writes
// out an Entry. Repeated use of Hooks is additive.
//
// Hooks are useful for simple side effects, like capturing metrics for the
// number of emitted logs. More complex side effects, including anything that
// requires access to the Entry's structured fields, should be implemented as
// a zapcore.Core instead. See zapcore.RegisterHooks for details.
func Hooks(hooks ...func(zapcore.Entry) error) Option {
	return optionFunc(func(log *Logger) {
		log.core = zapcore.RegisterHooks(log.core, hooks...)
	})
}

// Fields adds fields to the Logger.
func Fields(fs ...Field) Option {
	return optionFunc(func(log *Logger) {
		log.core = log.core.With(fs)
	})
}

// ErrorOutput sets the destination for errors generated by the Logger. Note
// that this option only affects internal errors; for sample code that sends
// error-level logs to a different location from info- and debug-level logs,
// see the package-level AdvancedConfiguration example.
//
// The supplied WriteSyncer must be safe for concurrent use. The Open and
// zapcore.Lock functions are the simplest ways to protect files with a mutex.
func ErrorOutput(w zapcore.WriteSyncer) Option {
	return optionFunc(func(log *Logger) {
		log.errorOutput = w
	})
}

// Development puts the logger in development mode, which makes DPanic-level
// logs panic instead of simply logging an error.
func Development() Option {
	return optionFunc(func(log *Logger) {
		log.development = true
	})
}

// AddCaller configures the Logger to annotate each message with the filename,
// line number, and function name of zap's caller. See also WithCaller.
func AddCaller() Option {
	return WithCaller(true)
}

// WithCaller configures the Logger to annotate each message with the filename,
// line number, and function name of zap's caller, or not, depending on the
// value of enabled. This is a generalized form of AddCaller.
func WithCaller(enabled bool) Option {
	return optionFunc(func(log *Logger) {
		log.addCaller = enabled
	})
}

// AddCallerSkip increases the number of callers skipped by caller annotation
// (as enabled by the AddCaller option). When building wrappers around the
// Logger and SugaredLogger, supplying this Option prevents zap from always
// reporting the wrapper code as the caller.
func AddCallerSkip(skip int) Option {
	return optionFunc(func(log *Logger) {
		log.callerSkip += skip
	})
}

// AddStacktrace configures the Logger to record a stack trace for all messages at
// or above a given level.
func AddStacktrace(lvl zapcore.LevelEnabler) Option {
	return optionFunc(func(log *Logger) {
		log.addStack = lvl
	})
}

// IncreaseLevel increase the level of the logger. It has no effect if
// the passed in level tries to decrease the level of the logger.
func IncreaseLevel(lvl zapcore.LevelEnabler) Option {
	return optionFunc(func(log *Logger) {
		core, err := zapcore.NewIncreaseLevelCore(log.core, lvl)
		if err != nil {
			_, _ = fmt.Fprintf(
				log.errorOutput,
				"failed to IncreaseLevel: %v\n",
				err,
			)
		} else {
			log.core = core
		}
	})
}

// WithPanicHook sets a CheckWriteHook to run on Panic/DPanic logs.
// Zap will call this hook after writing a log statement with a Panic/DPanic level.
//
// For example, the following builds a logger that will exit the current
// goroutine after writing a Panic/DPanic log message, but it will not start a panic.
//
//	zap.New(core, zap.WithPanicHook(zapcore.WriteThenGoexit))
//
// This is useful for testing Panic/DPanic log output.
func WithPanicHook(hook zapcore.CheckWriteHook) Option {
	return optionFunc(func(log *Logger) {
		log.onPanic = hook
	})
}

// OnFatal sets the action to take on fatal logs.
//
// Deprecated: Use [WithFatalHook] instead.
func OnFatal(action zapcore.CheckWriteAction) Option {
	return WithFatalHook(action)
}

// WithFatalHook sets a CheckWriteHook to run on fatal logs.
// Zap will call this hook after writing a log statement with a Fatal level.
//
// For example, the following builds a logger that will exit the current
// goroutine after writing a fatal log message, but it will not exit the
// program.
//
//	zap.New(core, zap.WithFatalHook(zapcore.WriteThenGoexit))
//
// It is important that the provided CheckWriteHook stops the control flow at
// the current statement to meet expectations of callers of the logger.
// We recommend calling os.Exit or runtime.Goexit inside custom hooks at
// minimum.
func WithFatalHook(hook zapcore.CheckWriteHook) Option {
	return optionFunc(func(log *Logger) {
		log.onFatal = hook
	})
}

// WithClock specifies the clock used by the logger to determine the current
// time for logged entries. Defaults to the system clock with time.Now.
func WithClock(clock zapcore.Clock) Option {
	return optionFunc(func(log *Logger) {
		log.clock = clock
	})
}
