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
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/uber-common/zap"
)

var _callersPool = sync.Pool{
	New: func() interface{} { return [1]uintptr{} },
}

// Sample returns a sampling logger. Each tick, the sampler records the first N
// messages from each call site and every Mth message thereafter. Sampling
// loggers are safe for concurrent use.
func Sample(zl zap.Logger, tick time.Duration, first, thereafter int) zap.Logger {
	return &sampler{
		Logger:     zl,
		tick:       tick,
		counts:     make(map[uintptr]*uint64),
		first:      uint64(first),
		thereafter: uint64(thereafter),
	}
}

type sampler struct {
	zap.Logger
	sync.RWMutex

	tick       time.Duration
	counts     map[uintptr]*uint64
	first      uint64
	thereafter uint64
}

func (s *sampler) With(fields ...zap.Field) zap.Logger {
	return &sampler{
		Logger:     s.Logger.With(fields...),
		tick:       s.tick,
		counts:     make(map[uintptr]*uint64),
		first:      s.first,
		thereafter: s.thereafter,
	}
}

func (s *sampler) Debug(msg string, fields ...zap.Field) {
	if s.check(zap.Debug) {
		s.Logger.Debug(msg, fields...)
	}
}

func (s *sampler) Info(msg string, fields ...zap.Field) {
	if s.check(zap.Info) {
		s.Logger.Info(msg, fields...)
	}
}

func (s *sampler) Warn(msg string, fields ...zap.Field) {
	if s.check(zap.Warn) {
		s.Logger.Warn(msg, fields...)
	}
}

func (s *sampler) Error(msg string, fields ...zap.Field) {
	if s.check(zap.Error) {
		s.Logger.Error(msg, fields...)
	}
}

func (s *sampler) Panic(msg string, fields ...zap.Field) {
	if s.check(zap.Panic) {
		s.Logger.Panic(msg, fields...)
	}
}

func (s *sampler) Fatal(msg string, fields ...zap.Field) {
	if s.check(zap.Fatal) {
		s.Logger.Fatal(msg, fields...)
	}
}

func (s *sampler) DFatal(msg string, fields ...zap.Field) {
	if s.check(zap.Error) {
		s.Logger.DFatal(msg, fields...)
	}
}

func (s *sampler) check(lvl zap.Level) bool {
	// If this log level isn't enabled, don't count this statement.
	if !s.Enabled(lvl) {
		return false
	}

	// Allocate this on the stack, since it doesn't escape.
	callers := [1]uintptr{}
	// Note that the skip argument for runtime.Callers is subtly different from
	// the skip argument for runtime.Caller.
	runtime.Callers(3, callers[:])
	caller := callers[0]

	s.RLock()
	if _, ok := s.counts[caller]; !ok {
		s.RUnlock()
		zero := uint64(0)
		s.Lock()
		s.counts[caller] = &zero
		s.Unlock()
		s.RLock()
	}
	count := s.counts[caller]
	s.RUnlock()
	n := atomic.AddUint64(count, 1)
	if n <= s.first {
		return true
	}
	if n == s.first+1 {
		// We've started sampling, reset the counter in a tick.
		time.AfterFunc(s.tick, func() { s.Reset(caller) })
	}
	return (n-s.first)%s.thereafter == 0
}

func (s *sampler) Reset(caller uintptr) {
	s.Lock()
	count := s.counts[caller]
	s.Unlock()
	atomic.StoreUint64(count, 0)
}
