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

package zap

import (
	"testing"

	"go.uber.org/atomic"
)

func benchmarkLoggers(levels ...Level) []Logger {
	logs := make([]Logger, len(levels))
	for i, lvl := range levels {
		logs[i] = Neo(WriterFacility(NullEncoder(), Discard, lvl))
	}
	return logs
}

func runIndexedPara(b *testing.B, f func(pb *testing.PB, j int)) {
	p := atomic.NewInt32(0)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		f(pb, int(p.Inc()))
	})
}

func BenchmarkCheckedMessage_Chain(b *testing.B) {
	logs := benchmarkLoggers(InfoLevel, ErrorLevel)
	data := []struct {
		lvl Level
		msg string
	}{
		{DebugLevel, "meh"},
		{InfoLevel, "fwiw"},
		{ErrorLevel, "hey!"},
	}
	runIndexedPara(b, func(pb *testing.PB, j int) {
		myInfoLog := logs[0].With(Int("p", j))
		myErrorLog := logs[1].With(Int("p", j))
		i := 0
		for pb.Next() {
			d := data[i%len(data)]
			cm := myInfoLog.Check(d.lvl, d.msg)
			cm = cm.Chain(myErrorLog.Check(d.lvl, d.msg))
			if cm.OK() {
				cm.Write(Int("i", i))
			}
			i++
		}
	})
}

func BenchmarkCheckedMessage_Chain_sliceLoggers(b *testing.B) {
	logs := benchmarkLoggers(InfoLevel, ErrorLevel)
	data := []struct {
		lvl Level
		msg string
	}{
		{DebugLevel, "meh"},
		{InfoLevel, "fwiw"},
		{ErrorLevel, "hey!"},
	}
	runIndexedPara(b, func(pb *testing.PB, j int) {
		myLogs := make([]Logger, len(logs))
		for i, log := range logs {
			myLogs[i] = log.With(Int("p", j))
		}
		i := 0
		for pb.Next() {
			d := data[i%len(data)]
			var cm *CheckedMessage
			for _, log := range myLogs {
				cm = cm.Chain(log.Check(d.lvl, d.msg))
			}
			if cm.OK() {
				cm.Write(Int("i", i))
			}
			i++
		}
	})
}
