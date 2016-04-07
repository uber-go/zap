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

import "runtime"

// takeStacktrace attempts to use the provided byte slice to take a stacktrace.
// If the provided slice isn't large enough, takeStacktrace will allocate
// successively larger slices until it can capture the whole stack.
func takeStacktrace(buf []byte, includeAllGoroutines bool) string {
	if len(buf) == 0 {
		// We may have been passed a nil byte slice.
		buf = make([]byte, 1024)
	}
	n := runtime.Stack(buf, includeAllGoroutines)
	for n >= len(buf) {
		// Buffer wasn't large enough, allocate a larger one. No need to copy
		// previous buffer's contents.
		size := 2 * n
		buf = make([]byte, size)
		n = runtime.Stack(buf, includeAllGoroutines)
	}
	return string(buf[:n])
}
