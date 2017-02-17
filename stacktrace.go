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
	"runtime"

	"go.uber.org/zap/buffer"
)

// takeStacktrace attempts to use the provided byte slice to take a stacktrace.
// If the provided slice isn't large enough, takeStacktrace will allocate
// successively larger slices until it can capture the whole stack.
func takeStacktrace(buffer *buffer.Buffer, skip int) string {
	size := 16
	programCounters := make([]uintptr, size)
	n := runtime.Callers(skip+2, programCounters)
	for n >= size {
		size *= 2
		programCounters = make([]uintptr, size)
		n = runtime.Callers(skip+2, programCounters)
	}
	programCounters = programCounters[:n]

	for i, programCounter := range programCounters {
		f := runtime.FuncForPC(programCounter)
		name := f.Name()
		// this strips everything below runtime.main
		// TODO(pedge): we might want to do a better check than this
		if name == "runtime.main" {
			break
		}
		// heh
		if i != 0 {
			buffer.AppendByte('\n')
		}
		file, line := f.FileLine(programCounter - 1)
		buffer.AppendString(f.Name())
		buffer.AppendByte('\n')
		buffer.AppendByte('\t')
		buffer.AppendString(file)
		buffer.AppendByte(':')
		buffer.AppendInt(int64(line))
	}
	return buffer.String()
}
