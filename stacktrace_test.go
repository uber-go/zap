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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTakeStacktrace(t *testing.T) {
	// Even if we pass a tiny buffer, takeStacktrace should allocate until it
	// can capture the whole stacktrace.
	traceNil := takeStacktrace(nil, false)
	traceTiny := takeStacktrace(make([]byte, 1), false)
	for _, trace := range []string{traceNil, traceTiny} {
		// The top frame should be takeStacktrace.
		assert.Contains(t, trace, "zap.takeStacktrace", "Stacktrace should contain the takeStacktrace function.")
		// The stacktrace should also capture its immediate caller.
		assert.Contains(t, trace, "TestTakeStacktrace", "Stacktrace should contain the test function.")
	}
}
