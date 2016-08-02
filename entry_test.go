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
	"time"

	"github.com/stretchr/testify/assert"
)

func stubNow(afterEpoch time.Duration) func() {
	prev := _timeNow
	t := time.Unix(0, int64(afterEpoch))
	_timeNow = func() time.Time { return t }
	return func() { _timeNow = prev }
}

func TestNewEntry(t *testing.T) {
	defer stubNow(0)()
	e := newEntry(DebugLevel, "hello", nil)
	assert.Equal(t, DebugLevel, e.Level, "Unexpected log level.")
	assert.Equal(t, time.Unix(0, 0).UTC(), e.Time, "Unexpected time.")
	assert.Nil(t, e.Fields(), "Unexpected fields.")
}
