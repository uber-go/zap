// Copyright (c) 2023 Uber Technologies, Inc.
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

package ztest

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMockClock_NewTicker(t *testing.T) {
	var n atomic.Int32
	clock := NewMockClock()

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
				n.Add(1)
			}
		}
	}(clock.NewTicker(time.Microsecond))

	// Move clock forward.
	clock.Add(2 * time.Microsecond)
	assert.Equal(t, int32(2), n.Load())
	close(quit)
}

func TestMockClock_NewTicker_slowConsumer(t *testing.T) {
	clock := NewMockClock()

	ticker := clock.NewTicker(time.Microsecond)
	defer ticker.Stop()

	// Two ticks, only one consumed.
	clock.Add(2 * time.Microsecond)
	<-ticker.C

	select {
	case <-ticker.C:
		t.Fatal("unexpected tick")
	default:
		// ok
	}
}

func TestMockClock_Add_negative(t *testing.T) {
	clock := NewMockClock()
	assert.Panics(t, func() { clock.Add(-1) })
}
