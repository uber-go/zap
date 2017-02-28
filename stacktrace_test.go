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
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	_stacktraceTestLock sync.Mutex
)

func TestEmptyStacktrace(t *testing.T) {
	withStacktraceIgnorePrefixes(
		[]string{
			"go.uber.org/zap",
			"runtime.",
			"testing.",
		},
		func() {
			assert.Empty(t, takeStacktrace(), "stacktrace should be empty if we ignore runtime, testing, and functions in this package")
		},
	)
}

func TestAllStacktrace(t *testing.T) {
	withNoStacktraceIgnorePrefixes(
		func() {
			stacktrace := takeStacktrace()
			assert.True(t, strings.HasPrefix(stacktrace, "go.uber.org/zap.TestAllStacktrace"),
				"stacktrace should start with TestAllStacktrace if no prefixes set, but is %s", stacktrace)
		},
	)
}

func withNoStacktraceIgnorePrefixes(f func()) {
	withStacktraceIgnorePrefixes([]string{}, f)
}

func withStacktraceIgnorePrefixes(prefixes []string, f func()) {
	_stacktraceTestLock.Lock()
	defer _stacktraceTestLock.Unlock()
	existing := _stacktraceIgnorePrefixes
	_stacktraceIgnorePrefixes = prefixes
	defer func() {
		_stacktraceIgnorePrefixes = existing
	}()
	f()
}
