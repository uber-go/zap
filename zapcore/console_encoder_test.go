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
package zapcore_test

import (
	"github.com/stretchr/testify/assert"
	. "go.uber.org/zap/zapcore"
	"testing"
)

func TestSetConsoleElementDelimiter(t *testing.T) {
	SetConsoleElementDelimiter(' ')
	enc := NewConsoleEncoder(humanEncoderConfig())
	enc.AddString("str", "foo")
	enc.AddInt64("int64-1", 1)

	buf, _ := enc.EncodeEntry(Entry{
		Message: "fake",
		Level:   DebugLevel,
	}, nil)

	assert.Equal(t, `0001-01-01T00:00:00.000Z DEBUG fake {"str": "foo", "int64-1": 1}
`, buf.String())
	buf.Free()
	SetConsoleElementDelimiter('\t')
}
