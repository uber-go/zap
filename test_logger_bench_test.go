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

	"github.com/uber-go/zap"
)

func withBenchedTee(b *testing.B, f func(zap.Logger)) {
	logger := zap.Tee(
		zap.New(
			zap.NewJSONEncoder(),
			zap.DebugLevel,
			zap.DiscardOutput,
		),
		zap.New(
			zap.NewJSONEncoder(),
			zap.InfoLevel,
			zap.DiscardOutput,
		),
	)
	b.ResetTimer()
	f(logger)
}

func BenchmarkTee_Check(b *testing.B) {
	cases := []struct {
		lvl zap.Level
		msg string
	}{
		{zap.DebugLevel, "foo"},
		{zap.InfoLevel, "bar"},
		{zap.WarnLevel, "baz"},
		{zap.ErrorLevel, "babble"},
	}
	withBenchedTee(b, func(logger zap.Logger) {
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				tt := cases[i]
				if cm := logger.Check(tt.lvl, tt.msg); cm.OK() {
					cm.Write(zap.Int("i", i))
				}
				i = (i + 1) % len(cases)
			}
		})
	})
}
