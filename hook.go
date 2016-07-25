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
	"path/filepath"
	"runtime"
	"strconv"
)

var (
	errHookNilEntry = errors.New("can't call a hook on a nil *Entry")
	// Skip Caller, Logger.log, and the leveled Logger method when using
	// runtime.Caller.
	_callerSkip = 3
)

// A hook is executed each time the logger writes a message. It can modify the
// message, but must not retain references to the message or any of its
// contents. Returned errors are written to the logger's error output.
type hook func(*Entry) error

// apply implements the Option interface.
func (h hook) apply(cfg *Config) {
	cfg.Hooks = append(cfg.Hooks, h)
}

// AddCaller configures the Logger to annotate each message with the filename
// and line number of zap's caller.
func AddCaller() Option {
	return hook(func(e *Entry) error {
		if e == nil {
			return errHookNilEntry
		}
		_, filename, line, ok := runtime.Caller(_callerSkip)
		if !ok {
			return errors.New("failed to get caller")
		}

		// Re-use a buffer from the pool.
		enc := newJSONEncoder()
		buf := enc.bytes
		buf = append(buf, filepath.Base(filename)...)
		buf = append(buf, ':')
		buf = strconv.AppendInt(buf, int64(line), 10)
		buf = append(buf, ':', ' ')
		buf = append(buf, e.Message...)

		newMsg := string(buf)
		enc.Free()
		e.Message = newMsg
		return nil
	})
}

// AddStacks configures the Logger to record a stack trace for all messages at
// or above a given level. Keep in mind that this is (relatively speaking) quite
// expensive.
func AddStacks(lvl Level) Option {
	return hook(func(e *Entry) error {
		if e == nil {
			return errHookNilEntry
		}
		if e.Level >= lvl {
			Stack().AddTo(e.Fields())
		}
		return nil
	})
}
