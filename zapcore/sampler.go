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

package zapcore

import (
	"time"
)

type sampler struct {
	Core

	counts            *counters
	tick              time.Duration
	first, thereafter uint64
}

// NewSampler creates a Core that samples incoming entries, which caps the CPU
// and I/O load of logging while attempting to preserve a representative subset
// of your logs.
//
// Zap samples by logging the first N entries with a given level and message
// each tick. If more Entries with the same level and message are seen during
// the same interval, every Mth message is logged and the rest are dropped.
//
// Keep in mind that zap's sampling implementation is optimized for speed over
// absolute precision; under load, each tick may be slightly over- or
// under-sampled.
func NewSampler(core Core, tick time.Duration, first, thereafter int) Core {
	return &sampler{
		Core:       core,
		tick:       tick,
		counts:     newCounters(),
		first:      uint64(first),
		thereafter: uint64(thereafter),
	}
}

func (s *sampler) With(fields []Field) Core {
	return &sampler{
		Core:       s.Core.With(fields),
		tick:       s.tick,
		counts:     s.counts,
		first:      s.first,
		thereafter: s.thereafter,
	}
}

func (s *sampler) Check(ent Entry, ce *CheckedEntry) *CheckedEntry {
	if !s.Enabled(ent.Level) {
		return ce
	}

	counter := s.counts.get(ent.Level, ent.Message)
	n, _ := counter.IncCheckReset(ent.Time, s.tick)
	if n > s.first && (n-s.first)%s.thereafter != 0 {
		return ce
	}
	return s.Core.Check(ent, ce)
}
