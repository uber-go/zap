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
	"unicode/utf8"

	"go.uber.org/zap/internal/bufferpool"

	"github.com/stretchr/testify/assert"
)

func TestTakeStacktrace(t *testing.T) {
	traceZero := takeStacktrace(bufferpool.Get(), 0)
	for _, trace := range []string{traceZero} {
		// The stacktrace should also its immediate caller.
		assert.Contains(t, trace, "TestTakeStacktrace", "Stacktrace should contain the test function.")
	}
}

func TestRunes(t *testing.T) {
	// https://golang.org/src/bytes/buffer.go?s=8208:8261#L237
	// I think this test might not be needed, because this might be checked
	// implicitly by the fact that we can pass these as bytes to AppendByte
	assert.True(t, '\n' < utf8.RuneSelf, `You can't cast '\n' to a byte, stop being silly`)
	assert.True(t, '\t' < utf8.RuneSelf, `You can't cast '\t' to a byte, stop being silly`)
	assert.True(t, ':' < utf8.RuneSelf, `You can't cast ':' to a byte, stop being silly`)
}
