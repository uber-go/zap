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
	"container/heap"
	"sync"
	"time"
)

// MockClock is a fake source of time.
// It implements standard time operations,
// but allows the user to control the passage of time.
//
// Use the [Add] method to progress time.
type MockClock struct {
	mu  sync.RWMutex
	now time.Time

	// The MockClock works by maintaining a list of waiters.
	// Each waiter knows the time at which it should be resolved.
	// When the clock advances, all waiters that are in range are resolved
	// in chronological order.
	waiters waiters
}

// NewMockClock builds a new mock clock
// using the current actual time as the initial time.
func NewMockClock() *MockClock {
	return &MockClock{
		now: time.Now(),
	}
}

// Now reports the current time.
func (c *MockClock) Now() time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.now
}

// NewTicker returns a time.Ticker that ticks at the specified frequency.
//
// As with [time.NewTicker],
// the ticker will drop ticks if the receiver is slow,
// and the channel is never closed.
//
// Calling Stop on the returned ticker is a no-op.
// The ticker only runs when the clock is advanced.
func (c *MockClock) NewTicker(d time.Duration) *time.Ticker {
	ch := make(chan time.Time, 1)

	var tick func(time.Time)
	tick = func(now time.Time) {
		next := now.Add(d)
		c.runAt(next, func() {
			defer tick(next)

			select {
			case ch <- next:
				// ok
			default:
				// The receiver is slow.
				// Drop the tick and continue.
			}
		})
	}
	tick(c.Now())

	return &time.Ticker{C: ch}
}

// Add progresses time by the given duration.
//
// Other operations waiting for the time to advance
// will be resolved if they are within range.
//
// Panics if the duration is negative.
func (c *MockClock) Add(d time.Duration) {
	if d < 0 {
		panic("cannot add negative duration")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	newTime := c.now.Add(d)
	// newTime won't be recorded until the end of this method.
	// This ensures that any waiters that are resolved
	// are resolved at the time they were expecting.

	for w, ok := c.waiters.PopLTE(newTime); ok; w, ok = c.waiters.PopLTE(newTime) {
		// The waiter is within range.
		// Travel to the time of the waiter and resolve it.
		c.now = w.until

		// The waiter may schedule more work
		// so we must release the lock.
		c.mu.Unlock()
		w.fn()
		// Sleeping here is necessary to let the side effects of waiters
		// take effect before we continue.
		time.Sleep(1 * time.Millisecond)
		c.mu.Lock()
	}

	c.now = newTime
}

// runAt schedules the given function to be run at the given time.
// The function runs without a lock held, so it may schedule more work.
func (c *MockClock) runAt(t time.Time, fn func()) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.waiters.Push(waiter{until: t, fn: fn})
}

type waiter struct {
	until time.Time
	fn    func()
}

// waiters is a thread-safe collection of waiters
// with the next waiter to be resolved at the front.
//
// Use the methods on this type to manipulate the collection.
// Do not modify the slice directly.
type waiters struct{ heap waiterHeap }

// Push adds a new waiter to the collection.
func (w *waiters) Push(v waiter) {
	heap.Push(&w.heap, v)
}

// PopLTE removes and returns the next waiter to be resolved
// if it is scheduled to be resolved at or before the given time.
//
// Returns false if there are no waiters in range.
func (w *waiters) PopLTE(t time.Time) (_ waiter, ok bool) {
	if len(w.heap) == 0 || w.heap[0].until.After(t) {
		return waiter{}, false
	}

	return heap.Pop(&w.heap).(waiter), true
}

// waiterHeap implements a min-heap of waiters based on their 'until' time.
//
// This is separate from the waiters type so that we can implement heap.Interface
// while still exposing a type-safe API on waiters.
type waiterHeap []waiter

var _ heap.Interface = (*waiterHeap)(nil)

func (h waiterHeap) Len() int      { return len(h) }
func (h waiterHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }

func (h waiterHeap) Less(i, j int) bool {
	return h[i].until.Before(h[j].until)
}

func (h *waiterHeap) Push(x interface{}) {
	*h = append(*h, x.(waiter))
}

func (h *waiterHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[:n-1]
	return x
}
