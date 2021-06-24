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
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/atomic"
)

// controlledClock provides control over the time via a mock clock.
type controlledClock struct{ *clock.Mock }

func newControlledClock() *controlledClock {
	return &controlledClock{clock.NewMock()}
}

func (c *controlledClock) NewTicker(d time.Duration) *time.Ticker {
	return &time.Ticker{C: c.Ticker(d).C}
}

func TestControlledClock_NewTicker(t *testing.T) {
	var n atomic.Int32
	ctrlMock := newControlledClock()

	done := make(chan struct{})
	defer func() { <-done }() // wait for end

	quit := make(chan struct{})
	// Create a channel to increment every microsecond.
	go func(ticker *time.Ticker) {
		defer close(done)
		for {
			select {
			case <-quit:
				ticker.Stop()
				return
			case <-ticker.C:
				n.Inc()
			}
		}
	}(ctrlMock.NewTicker(time.Microsecond))

	// Move clock forward.
	ctrlMock.Add(2 * time.Microsecond)
	assert.Equal(t, int32(2), n.Load())
	close(quit)
}

func TestSystemClock_NewTicker(t *testing.T) {
	want := 3

	var n int
	timer := DefaultClock.NewTicker(time.Millisecond)
	for range timer.C {
		n++
		if n == want {
			return
		}
	}
}
