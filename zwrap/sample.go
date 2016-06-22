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

	"github.com/uber-go/zap"

	"github.com/uber-go/atomic"
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

// Sample returns a sampling logger. The logger maintains a separate bucket
// for each message (e.g., "foo" in logger.Warn("foo")). In each tick, the
// sampler will emit the first N logs in each bucket and every Mth log
// therafter. Sampling loggers are safe for concurrent use.
//
// Per-message counts are shared between parent and child loggers, which allows
// applications to more easily control global I/O load.
func Sample(zl zap.Logger, tick time.Duration, first, thereafter int) zap.Logger {
	return &sampler{
		Logger:     zl,
		tick:       tick,
		counts:     &counters{counts: make(map[string]*atomic.Uint64)},
		first:      uint64(first),
		thereafter: uint64(thereafter),
	}
}

type sampler struct {
	zap.Logger

	tick       time.Duration
	counts     *counters
	first      uint64
	thereafter uint64
}

func (s *sampler) With(fields ...zap.Field) zap.Logger {
	return &sampler{
		Logger:     s.Logger.With(fields...),
		tick:       s.tick,
		counts:     s.counts,
		first:      s.first,
		thereafter: s.thereafter,
	}
}

func (s *sampler) Check(lvl zap.Level, msg string) *zap.CheckedMessage {
	if !s.check(lvl, msg) {
		return nil
	}
	return zap.NewCheckedMessage(s.Logger, lvl, msg)
}

func (s *sampler) Log(lvl zap.Level, msg string, fields ...zap.Field) {
	if s.check(lvl, msg) {
		s.Logger.Log(lvl, msg, fields...)
	}
}

func (s *sampler) Debug(msg string, fields ...zap.Field) {
	if s.check(zap.DebugLevel, msg) {
		s.Logger.Debug(msg, fields...)
	}
}

func (s *sampler) Info(msg string, fields ...zap.Field) {
	if s.check(zap.InfoLevel, msg) {
		s.Logger.Info(msg, fields...)
	}
}

func (s *sampler) Warn(msg string, fields ...zap.Field) {
	if s.check(zap.WarnLevel, msg) {
		s.Logger.Warn(msg, fields...)
	}
}

func (s *sampler) Error(msg string, fields ...zap.Field) {
	if s.check(zap.ErrorLevel, msg) {
		s.Logger.Error(msg, fields...)
	}
}

func (s *sampler) Panic(msg string, fields ...zap.Field) {
	if s.check(zap.PanicLevel, msg) {
		s.Logger.Panic(msg, fields...)
	}
}

func (s *sampler) Fatal(msg string, fields ...zap.Field) {
	if s.check(zap.FatalLevel, msg) {
		s.Logger.Fatal(msg, fields...)
	}
}

func (s *sampler) DFatal(msg string, fields ...zap.Field) {
	if s.check(zap.ErrorLevel, msg) {
		s.Logger.DFatal(msg, fields...)
	}
}

func (s *sampler) check(lvl zap.Level, msg string) bool {
	if !(lvl >= s.Level()) {
		return false
	}
	n := s.counts.Inc(msg)
	if n <= s.first {
		return true
	}
	if n == s.first+1 {
		time.AfterFunc(s.tick, func() { s.counts.Reset(msg) })
	}
	return (n-s.first)%s.thereafter == 0
}
