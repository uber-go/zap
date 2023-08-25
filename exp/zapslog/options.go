// Copyright (c) 2023 Uber Technologies, Inc.
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

package zapslog

import "go.uber.org/zap/zapcore"

type Option interface {
	apply(*Handler)
}

// optionFunc wraps a func so it satisfies the Option interface.
type optionFunc func(*Handler)

func (f optionFunc) apply(handler *Handler) {
	f(handler)
}

// WithName configures the Logger to annotate each message with the logger name.
func WithName(name string) Option {
	return optionFunc(func(h *Handler) {
		h.name = name
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
	return optionFunc(func(handler *Handler) {
		handler.addCaller = enabled
	})
}

// AddCallerSkip increases the number of callers skipped by caller annotation
// (as enabled by the AddCaller option). When building wrappers around the
// Logger and SugaredLogger, supplying this Option prevents zap from always
// reporting the wrapper code as the caller.
func AddCallerSkip(skip int) Option {
	return optionFunc(func(log *Handler) {
		log.callerSkip += skip
	})
}

// AddStacktrace configures the Logger to record a stack trace for all messages at
// or above a given level.
func AddStacktrace(lvl zapcore.LevelEnabler) Option {
	return optionFunc(func(log *Handler) {
		log.addStack = lvl
	})
}
