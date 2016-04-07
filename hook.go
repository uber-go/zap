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

// Skip Caller, Logger.log, and the leveled Logger method when using
// runtime.Caller.
var _callerSkip = 3

// A hook is executed each time the Logger writes a message. It receives the
// message's priority, the message itself, and the logging context, and it
// returns a modified message and an error.
type hook func(Level, string, KeyValue) (string, error)

// apply implements the Option interface.
func (h hook) apply(jl *jsonLogger) error {
	jl.hooks = append(jl.hooks, h)
	return nil
}

// AddCaller configures the Logger to annotate each message with the filename
// and line number of zap's caller.
func AddCaller() Option {
	return hook(func(_ Level, msg string, _ KeyValue) (string, error) {
		_, filename, line, ok := runtime.Caller(_callerSkip)
		if !ok {
			return msg, errors.New("failed to get caller")
		}

		// Re-use a buffer from the pool.
		enc := newJSONEncoder()
		buf := enc.bytes
		buf = append(buf, filepath.Base(filename)...)
		buf = append(buf, ':')
		buf = strconv.AppendInt(buf, int64(line), 10)
		buf = append(buf, ':', ' ')
		buf = append(buf, msg...)

		newMsg := string(buf)
		enc.Free()
		return newMsg, nil
	})
}

// AddStacks configures the Logger to record a stack trace for all messages at
// or above a given level. Keep in mind that this is (relatively speaking) quite
// expensive.
func AddStacks(lvl Level) Option {
	return hook(func(msgLevel Level, msg string, kv KeyValue) (string, error) {
		if msgLevel >= lvl {
			Stack().addTo(kv)
		}
		return msg, nil
	})
}
