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
	"strings"
	"sync"

	"go.uber.org/zap/internal/bufferpool"
)

var (
	_stacktraceIgnorePrefixes = []string{
		"go.uber.org/zap",
		"runtime.",
	}
	_stacktraceUintptrPool = sync.Pool{
		New: func() interface{} {
			return make([]uintptr, 64)
		},
	}
)

// takeStacktrace attempts to use the provided byte slice to take a stacktrace.
// If the provided slice isn't large enough, takeStacktrace will allocate
// successively larger slices until it can capture the whole stack.
func takeStacktrace() string {
	buffer := bufferpool.Get()
	programCounters := _stacktraceUintptrPool.Get().([]uintptr)
	defer bufferpool.Put(buffer)
	defer _stacktraceUintptrPool.Put(programCounters)

	for {
		n := runtime.Callers(2, programCounters)
		if n < len(programCounters) {
			programCounters = programCounters[:n]
			break
		}
		// Do not put programCounters back in pool, will put larger buffer in
		// This is in case our size is always wrong, we optimize to pul
		// correctly-sized buffers back in the pool
		programCounters = make([]uintptr, len(programCounters)*2)
	}

	i := 0
	for _, programCounter := range programCounters {
		f := runtime.FuncForPC(programCounter)
		name := f.Name()
		if shouldIgnoreStacktraceName(name) {
			continue
		}
		if i != 0 {
			buffer.AppendByte('\n')
		}
		i++
		file, line := f.FileLine(programCounter - 1)
		buffer.AppendString(name)
		buffer.AppendByte('\n')
		buffer.AppendByte('\t')
		buffer.AppendString(file)
		buffer.AppendByte(':')
		buffer.AppendInt(int64(line))
	}

	return buffer.String()
}

func shouldIgnoreStacktraceName(name string) bool {
	for _, prefix := range _stacktraceIgnorePrefixes {
		if strings.HasPrefix(name, prefix) {
			return true
		}
	}
	return false
}
