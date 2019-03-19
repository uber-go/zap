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
)

const (
	// ReportingMessage is the message in sampling report log entry.
	ReportingMessage = "Log entries were sampled"
	// ReportingLoggerName is the name field of sampling report log entry.
	ReportingLoggerName = "zapcore"
)

type reportingSampler struct {
	*sampler

	reportingCore Core
}

// NewReportingSampler creates a Core that samples incoming entries, which caps
// the CPU and I/O load of logging while attempting to preserve a representative
// subset of your logs. The difference with the regular sampling core is that it
// reports the number of sampled entries for the same logging level and message.
// Reporting is in form of a separate INFO log entry per every tick there were
// entries sampled. Therefore, core must have at least info logging level.
func NewReportingSampler(core Core, tick time.Duration, first, thereafter int) Core {
	if !core.Enabled(InfoLevel) {
		panic("Core must have at least info logging level.")
	}
	return &reportingSampler{
		sampler: &sampler{
			Core:       core,
			tick:       tick,
			counts:     newCounters(),
			first:      uint64(first),
			thereafter: uint64(thereafter),
		},
		reportingCore: core,
	}
}

func (rs *reportingSampler) With(fields []Field) Core {
	return &reportingSampler{
		sampler: &sampler{
			Core:       rs.Core.With(fields),
			tick:       rs.tick,
			counts:     rs.counts,
			first:      rs.first,
			thereafter: rs.thereafter,
		},
		// reportingCore to be used only for reporting,
		// we want to avoid extra fields there.
		reportingCore: rs.reportingCore,
	}
}

func (rs *reportingSampler) report(ent Entry, n uint64) {
	sampledCount := n - rs.first
	sampledCount = sampledCount - sampledCount/rs.thereafter
	if sampledCount == 0 {
		return
	}
	entry := Entry{
		Level:      InfoLevel,
		Time:       time.Now(),
		LoggerName: ReportingLoggerName,
		Message:    ReportingMessage,
	}
	fields := []Field{
		{Key: "original_level", Type: StringType, String: ent.Level.String()},
		{Key: "original_message", Type: StringType, String: ent.Message},
		{Key: "count", Type: Int64Type, Integer: int64(sampledCount)},
	}
	err := rs.reportingCore.Write(entry, fields)
	if err != nil {
		fmt.Printf("%v write error: %v\n", time.Now(), err)
	}
}

func (rs *reportingSampler) Check(ent Entry, ce *CheckedEntry) *CheckedEntry {
	if !rs.Enabled(ent.Level) {
		return ce
	}

	counter := rs.counts.get(ent.Level, ent.Message)
	n, prevN := counter.IncCheckReset(ent.Time, rs.tick)
	if n > rs.first && (n-rs.first)%rs.thereafter != 0 {
		return ce
	}
	if prevN > 1 {
		rs.report(ent, prevN)
	}
	return rs.Core.Check(ent, ce)
}
