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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTakeStacktrace(t *testing.T) {
	trace := takeStacktrace(0)
	lines := strings.Split(trace, "\n")
	require.NotEmpty(t, lines, "Expected stacktrace to have at least one frame.")
	assert.Contains(
		t,
		lines[0],
		"go.uber.org/zap.TestTakeStacktrace",
		"Expected stacktrace to start with the test.",
	)
}

func TestTakeStacktraceWithSkip(t *testing.T) {
	trace := takeStacktrace(1)
	lines := strings.Split(trace, "\n")
	require.NotEmpty(t, lines, "Expected stacktrace to have at least one frame.")
	assert.Contains(
		t,
		lines[0],
		"testing.",
		"Expected stacktrace to start with the test runner (skipping our own frame).",
	)
}

func TestTakeStacktraceWithSkipInnerFunc(t *testing.T) {
	var trace string
	func() {
		trace = takeStacktrace(2)
	}()
	lines := strings.Split(trace, "\n")
	require.NotEmpty(t, lines, "Expected stacktrace to have at least one frame.")
	assert.Contains(
		t,
		lines[0],
		"testing.",
		"Expected stacktrace to start with the test function (skipping the test function).",
	)
}

func BenchmarkTakeStacktrace(b *testing.B) {
	for i := 0; i < b.N; i++ {
		takeStacktrace(0)
	}
}
