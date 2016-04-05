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

package benchmarks

import (
	"io/ioutil"
	"testing"
	"time"

	"gopkg.in/inconshreveable/log15.v2"
)

func newLog15() log15.Logger {
	logger := log15.New()
	logger.SetHandler(log15.StreamHandler(ioutil.Discard, log15.JsonFormat()))
	return logger
}

func BenchmarkLog15AddingFields(b *testing.B) {
	logger := newLog15()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info("Go fast.",
				"one", 1,
				"two", int64(1),
				"three", 3.0,
				"four", "four!",
				"five", true,
				"six", time.Unix(0, 0),
				"error", errExample.Error(),
				"eight", time.Second,
				"nine", _jane,
				"ten", "done!",
			)
		}
	})
}

func BenchmarkLog15WithAccumulatedContext(b *testing.B) {
	logger := newLog15().New(
		"one", 1,
		"two", int64(1),
		"three", 3.0,
		"four", "four!",
		"five", true,
		"six", time.Unix(0, 0),
		"error", errExample.Error(),
		"eight", time.Second,
		"nine", _jane,
		"ten", "done!",
	)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info("Go really fast.")
		}
	})
}

func BenchmarkLog15WithoutFields(b *testing.B) {
	logger := newLog15()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info("Go fast.")
		}
	})
}
