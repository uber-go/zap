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
	"sync/atomic"
	"time"

	"github.com/uber-common/zap"
)

// Sample returns a sampling logger. Each tick, the sampler records the first N
// messages and every Mth message thereafter. Sampling loggers are safe for
// concurrent use.
//
// The children of sampling loggers (i.e., the results of calls to the With method)
// inherit their parent's sampling logic but maintain independent counters.
func Sample(zl zap.Logger, tick time.Duration, first, thereafter int) zap.Logger {
	return &sampler{
		Logger:     zl,
		tick:       tick,
		first:      uint64(first),
		thereafter: uint64(thereafter),
	}
}

type sampler struct {
	zap.Logger

	tick       time.Duration
	count      uint64
	first      uint64
	thereafter uint64
}

func (s *sampler) With(fields ...zap.Field) zap.Logger {
	return &sampler{
		Logger:     s.Logger.With(fields...),
		tick:       s.tick,
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
	n := atomic.AddUint64(&s.count, 1)
	if n <= s.first {
		return true
	}
	if n == s.first+1 {
		// We've started sampling, reset the counter in a tick.
		time.AfterFunc(s.tick, func() { s.Reset() })
	}
	return (n-s.first)%s.thereafter == 0
}

func (s *sampler) Reset() {
	atomic.StoreUint64(&s.count, 0)
}
