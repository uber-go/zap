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

const _zapPackage = "go.uber.org/zap"

var (
	// We add "." and "/" suffixes to the package name to ensure we only match
	// the exact package and not any package with the same prefix.
	_zapStacktracePrefixes       = addPrefix(_zapPackage, ".", "/")
	_zapStacktraceVendorContains = addPrefix("/vendor/", _zapStacktracePrefixes...)
)

func takeStacktrace() string {
	pcs := newProgramCounters()
	defer pcs.Release()

	return makeStacktrace(pcs.Callers(1))
}

func makeStacktrace(pcs []uintptr) string {
	buffer := bufferpool.Get()
	defer buffer.Free()

	i := 0
	skipZapFrames := true // skip all consecutive zap frames at the beginning.
	frames := runtime.CallersFrames(pcs)

	// Note: On the last iteration, frames.Next() returns false, with a valid
	// frame, but we ignore this frame. The last frame is a a runtime frame which
	// adds noise, since it's only either runtime.main or runtime.goexit.
	for frame, more := frames.Next(); more; frame, more = frames.Next() {
		if skipZapFrames && isZapFrame(frame.Function) {
			continue
		} else {
			skipZapFrames = false
		}

		if i != 0 {
			buffer.AppendByte('\n')
		}
		i++
		buffer.AppendString(frame.Function)
		buffer.AppendByte('\n')
		buffer.AppendByte('\t')
		buffer.AppendString(frame.File)
		buffer.AppendByte(':')
		buffer.AppendInt(int64(frame.Line))
	}

	return buffer.String()
}

func isZapFrame(function string) bool {
	for _, prefix := range _zapStacktracePrefixes {
		if strings.HasPrefix(function, prefix) {
			return true
		}
	}

	// We can't use a prefix match here since the location of the vendor
	// directory affects the prefix. Instead we do a contains match.
	for _, contains := range _zapStacktraceVendorContains {
		if strings.Contains(function, contains) {
			return true
		}
	}

	return false
}

type programCounters struct {
	pcs []uintptr
}

var _stacktracePool = sync.Pool{
	New: func() interface{} {
		return &programCounters{make([]uintptr, 64)}
	},
}

func newProgramCounters() *programCounters {
	return _stacktracePool.Get().(*programCounters)
}

func (pcs *programCounters) Release() {
	_stacktracePool.Put(pcs)
}

func (pcs *programCounters) Callers(skip int) []uintptr {
	var numFrames int
	for {
		// Skip the call to runtime.Callers and programCounter.Callers
		// so that the program counters start at our caller.
		numFrames = runtime.Callers(skip+2, pcs.pcs)
		if numFrames < len(pcs.pcs) {
			return pcs.pcs[:numFrames]
		}

		// Don't put the too-short counter slice back into the pool; this lets
		// the pool adjust if we consistently take deep stacktraces.
		pcs.pcs = make([]uintptr, len(pcs.pcs)*2)
	}
}

func addPrefix(prefix string, ss ...string) []string {
	withPrefix := make([]string, len(ss))
	for i, s := range ss {
		withPrefix[i] = prefix + s
	}
	return withPrefix
}
