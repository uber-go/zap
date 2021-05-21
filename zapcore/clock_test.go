// Copyright (c) 2021 Uber Technologies, Inc.
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

package zapcore

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/stretchr/testify/assert"
)

type controlledClock struct {
	mockClock *clock.Mock
}

func (c *controlledClock) Now() time.Time {
	return c.mockClock.Now()
}

func (c *controlledClock) NewTicker(d time.Duration) *time.Ticker {
	return &time.Ticker{C: c.mockClock.Ticker(d).C}
}

func TestMockClock(t *testing.T) {
	var n int32
	ctrlMock := &controlledClock{mockClock: clock.NewMock()}

	// Create a channel to increment every microsecond.
	go func() {
		ticker := ctrlMock.NewTicker(1 * time.Microsecond)
		for {
			<-ticker.C
			atomic.AddInt32(&n, 1)
		}
	}()
	time.Sleep(1 * time.Microsecond)

	// Move clock forward.
	ctrlMock.mockClock.Add(10 * time.Microsecond)
	assert.Equal(t, atomic.LoadInt32(&n), int32(10))
}
