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

package zwrap

import (
	"sync"
	"time"

	"go.uber.org/atomic"
	"go.uber.org/zap"
)

type counters struct {
	sync.RWMutex
	counts map[string]*atomic.Uint64
}

func (c *counters) Inc(key string) uint64 {
	c.RLock()
	count, ok := c.counts[key]
	c.RUnlock()
	if ok {
		return count.Inc()
	}

	c.Lock()
	count, ok = c.counts[key]
	if ok {
		c.Unlock()
		return count.Inc()
	}

	c.counts[key] = atomic.NewUint64(1)
	c.Unlock()
	return 1
}

func (c *counters) Reset(key string) {
	c.Lock()
	count := c.counts[key]
	c.Unlock()
	count.Store(0)
}

// Sample creates a sampled facility.
func Sample(fac zap.Facility, tick time.Duration, first, thereafter int) zap.Facility {
	return &sampler{
		Facility:   fac,
		tick:       tick,
		counts:     &counters{counts: make(map[string]*atomic.Uint64)},
		first:      uint64(first),
		thereafter: uint64(thereafter),
	}
}

type sampler struct {
	zap.Facility

	tick       time.Duration
	counts     *counters
	first      uint64
	thereafter uint64
}

func (s *sampler) With(fields ...zap.Field) zap.Facility {
	return &sampler{
		Facility:   s.Facility.With(fields...),
		tick:       s.tick,
		counts:     s.counts,
		first:      s.first,
		thereafter: s.thereafter,
	}
}

func (s *sampler) Log(ent zap.Entry, fields []zap.Field) error {
	if s.sampled(ent) {
		return s.Facility.Log(ent, fields)
	}
	return nil
}

func (s *sampler) Check(ent zap.Entry, ce *zap.CheckedEntry) *zap.CheckedEntry {
	if s.sampled(ent) {
		ce = s.Facility.Check(ent, ce)
	}
	return ce
}

func (s *sampler) sampled(ent zap.Entry) bool {
	if !s.Facility.Enabled(ent.Level) {
		return false
	}
	n := s.counts.Inc(ent.Message)
	if n <= s.first {
		return true
	}
	if n == s.first+1 {
		time.AfterFunc(s.tick, func() { s.counts.Reset(ent.Message) })
	}
	if (n-s.first)%s.thereafter == 0 {
		return true
	}
	return false
}
