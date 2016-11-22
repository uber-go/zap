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

package zap_test

import (
	"testing"

	"github.com/uber-go/atomic"
	"github.com/uber-go/zap"
)

func BenchmarkCheckedMessage_Chain(b *testing.B) {
	infoLog := zap.New(
		zap.NullEncoder(),
		zap.InfoLevel,
		zap.DiscardOutput,
	)
	errorLog := zap.New(
		zap.NullEncoder(),
		zap.ErrorLevel,
		zap.DiscardOutput,
	)

	data := []struct {
		lvl zap.Level
		msg string
	}{
		{zap.DebugLevel, "meh"},
		{zap.InfoLevel, "fwiw"},
		{zap.ErrorLevel, "hey!"},
	}

	p := atomic.NewInt32(0)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		j := p.Inc()
		myInfoLog := infoLog.With(zap.Int("p", int(j)))
		myErrorLog := errorLog.With(zap.Int("p", int(j)))
		for pb.Next() {
			d := data[i%len(data)]
			cm := myInfoLog.Check(d.lvl, d.msg)
			cm = cm.Chain(myErrorLog.Check(d.lvl, d.msg))
			if cm.OK() {
				cm.Write(zap.Int("i", i))
			}
			i++
		}
	})
}

func BenchmarkCheckedMessage_Chain_sliceLoggers(b *testing.B) {
	logs := []zap.Logger{
		zap.New(
			zap.NullEncoder(),
			zap.InfoLevel,
			zap.DiscardOutput,
		),
		zap.New(
			zap.NullEncoder(),
			zap.ErrorLevel,
			zap.DiscardOutput,
		),
	}

	data := []struct {
		lvl zap.Level
		msg string
	}{
		{zap.DebugLevel, "meh"},
		{zap.InfoLevel, "fwiw"},
		{zap.ErrorLevel, "hey!"},
	}

	p := atomic.NewInt32(0)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		j := p.Inc()
		myLogs := make([]zap.Logger, len(logs))
		for i, log := range logs {
			myLogs[i] = log.With(zap.Int("p", int(j)))
		}
		i := 0
		for pb.Next() {
			d := data[i%len(data)]
			var cm *zap.CheckedMessage
			for _, log := range myLogs {
				cm = cm.Chain(log.Check(d.lvl, d.msg))
			}
			if cm.OK() {
				cm.Write(zap.Int("i", i))
			}
			i++
		}
	})
}
