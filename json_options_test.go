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

func TestMessageFormatters(t *testing.T) {
	const msg = "foo"

	tests := []struct {
		name      string
		formatter MessageFormatter
		expected  Field
	}{
		{"MessageKey", MessageKey("the-message"), String("the-message", msg)},
		{"Default", defaultMessageF, String("msg", msg)},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, tt.formatter(msg), "Unexpected output from MessageFormatter %s.", tt.name)
	}
}

func TestTimeFormatters(t *testing.T) {
	tests := []struct {
		name      string
		formatter TimeFormatter
		expected  Field
	}{
		{"EpochFormatter", EpochFormatter("the-time"), Float64("the-time", 0)},
		{"RFC3339", RFC3339Formatter("ts"), String("ts", "1970-01-01T00:00:00Z")},
		{"NoTime", NoTime(), Skip()},
		{"Default", defaultTimeF, Float64("ts", 0)},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, tt.formatter(epoch), "Unexpected output from TimeFormatter %s.", tt.name)
	}
}

func TestLevelFormatters(t *testing.T) {
	const lvl = InfoLevel
	tests := []struct {
		name      string
		formatter LevelFormatter
		expected  Field
	}{
		{"LevelString", LevelString("the-level"), String("the-level", "info")},
		{"Default", defaultLevelF, String("level", "info")},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, tt.formatter(lvl), "Unexpected output from LevelFormatter %s.", tt.name)
	}
}
