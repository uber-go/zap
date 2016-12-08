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
	"errors"
	"runtime"
)

var (
	errHookNilEntry = errors.New("can't call a hook on a nil *Entry")
	errCaller       = errors.New("failed to get caller")
	// Skip Caller, Logger.log, and the leveled Logger method when using
	// runtime.Caller.
	_callerSkip = 3
)

// A Hook is executed each time the logger writes an Entry. It can modify the
// entry, but must not retain references to the entry or any of its contents.
// Returned errors are written to the logger's error output.
//
// Hooks implement the Option interface.
type Hook func(*Entry) error

// apply implements the Option interface.
func (h Hook) apply(log *logger) {
	log.hooks = append(log.hooks, h)
}

// AddCaller configures the Logger to annotate each message with the filename
// and line number of zap's caller.
func AddCaller() Option {
	return Hook(func(e *Entry) error {
		if e == nil {
			return errHookNilEntry
		}
		e.Caller = MakeEntryCaller(runtime.Caller(_callerSkip))
		if !e.Caller.Defined {
			return errCaller
		}
		return nil
	})
}

// AddStacks configures the Logger to record a stack trace for all messages at
// or above a given level. Keep in mind that this is (relatively speaking) quite
// expensive.
//
// TODO: why is this called AddStacks rather than just AddStack or AddStacktrace?
func AddStacks(lvl Level) Option {
	return Hook(func(e *Entry) error {
		if e == nil {
			return errHookNilEntry
		}
		if e.Level >= lvl {
			e.Stack = Stack().str
		}
		return nil
	})
}
