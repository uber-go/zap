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
	"fmt"
	"time"

	"go.uber.org/atomic"
)

const (
	_numLevels        = _maxLevel - _minLevel + 1
	_countersPerLevel = 4096

	// ReportingLoggerName is the default name of logger that generates sampling
	// reports.
	ReportingLoggerName = "zapcore"
	// ReportingLogMessage is the default message of a sampling report entry.
	ReportingLogMessage = "Log entries were sampled"
)

type counter struct {
	resetAt atomic.Int64
	counter atomic.Uint64
}

type counters [_numLevels][_countersPerLevel]counter

func newCounters() *counters {
	return &counters{}
}

func (cs *counters) get(lvl Level, key string) *counter {
	i := lvl - _minLevel
	j := fnv32a(key) % _countersPerLevel
	return &cs[i][j]
}

// fnv32a, adapted from "hash/fnv", but without a []byte(string) alloc
func fnv32a(s string) uint32 {
	const (
		offset32 = 2166136261
		prime32  = 16777619
	)
	hash := uint32(offset32)
	for i := 0; i < len(s); i++ {
		hash ^= uint32(s[i])
		hash *= prime32
	}
	return hash
}

func (c *counter) IncCheckReset(t time.Time, tick time.Duration) (currentVal, preResetVal uint64) {
	tn := t.UnixNano()
	resetAfter := c.resetAt.Load()
	if resetAfter > tn {
		return c.counter.Inc(), 0
	}

	newResetAfter := tn + tick.Nanoseconds()
	if !c.resetAt.CAS(resetAfter, newResetAfter) {
		// We raced with another goroutine trying to reset, and it reset the
		// counter to 1, so we just increment.
		return c.counter.Inc(), 0
	}
	preResetVal = c.counter.Load()
	currentVal = 1
	c.counter.Store(currentVal)
	return

}

type sampler struct {
	Core

	*reporting
	counts            *counters
	tick              time.Duration
	first, thereafter uint64
}

type reporting struct {
	enabled    bool
	core       Core
	loggerName string
	level      Level
	message    string
}

// NewReportingSampler creates a Core that samples incoming entries, which caps
// the CPU and I/O load of logging while attempting to preserve a representative
// subset of your logs. The difference with the regular sampling core is that it
// reports the number of sampled entries for the same logging level and message.
// Reporting is in form of a separate log entry per every tick when there are
// sampled entries. reportingLevel must be enabled on the core.
func NewReportingSampler(
	core Core,
	tick time.Duration,
	first, thereafter int,
	reportingLevel Level,
	reportingLoggerName, reportingMessage string,
) Core {
	if !core.Enabled(reportingLevel) {
		panic(fmt.Sprintf(
			"Core must have at least %s logging level.",
			reportingLevel.String(),
		))
	}
	if reportingLoggerName == "" {
		reportingLoggerName = ReportingLoggerName
	}
	if reportingMessage == "" {
		reportingMessage = ReportingLogMessage
	}
	reporting := &reporting{
		enabled:    true,
		core:       core,
		loggerName: reportingLoggerName,
		level:      reportingLevel,
		message:    reportingMessage,
	}
	return &sampler{
		Core:       core,
		reporting:  reporting,
		tick:       tick,
		counts:     newCounters(),
		first:      uint64(first),
		thereafter: uint64(thereafter),
	}
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
	reporting := &reporting{
		enabled: false,
		core:    NewNopCore(),
	}
	return &sampler{
		Core:       core,
		reporting:  reporting,
		tick:       tick,
		counts:     newCounters(),
		first:      uint64(first),
		thereafter: uint64(thereafter),
	}
}

func (s *sampler) With(fields []Field) Core {
	return &sampler{
		Core: s.Core.With(fields),
		// reporting.core is only to be used for sampling activity reporting.
		// No extra fields should be added to it.
		reporting:  s.reporting,
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
	n, prevN := counter.IncCheckReset(ent.Time, s.tick)
	if n > s.first && (n-s.first)%s.thereafter != 0 {
		return ce
	}
	if s.reporting.enabled && prevN > 0 {
		s.report(ent, prevN)
	}
	return s.Core.Check(ent, ce)
}

func (s *sampler) report(ent Entry, n uint64) error {
	sampleCount := n - s.first
	sampleCount = sampleCount - sampleCount/s.thereafter
	if sampleCount <= 0 {
		return nil
	}
	entry := Entry{
		LoggerName: s.reporting.loggerName,
		Time:       time.Now(),
		Level:      s.reporting.level,
		Message:    s.reporting.message,
	}
	fields := []Field{
		{Key: "original_level", Type: StringType, String: ent.Level.String()},
		{Key: "original_message", Type: StringType, String: ent.Message},
		{Key: "count", Type: Int64Type, Integer: int64(sampleCount)},
	}
	return s.reporting.core.Write(entry, fields)
}
