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

	"go.pedge.io/lion"
)

func BenchmarkLionAddingFieldsJSON(b *testing.B) {
	benchmarkLionAddingFields(b, newLionJSON())
}

func BenchmarkLionAddingFieldsText(b *testing.B) {
	benchmarkLionAddingFields(b, newLionText())
}

func benchmarkLionAddingFields(b *testing.B, logger lion.Logger) {
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.WithFields(map[string]interface{}{
				"int":               1,
				"int64":             int64(1),
				"float":             3.0,
				"string":            "four!",
				"bool":              true,
				"time":              time.Unix(0, 0),
				"error":             errExample.Error(),
				"duration":          time.Second,
				"user-defined type": _jane,
				"another string":    "done!",
			}).Infoln("Go fast.")
		}
	})
}

func BenchmarkLionWithAccumulatedContextJSON(b *testing.B) {
	benchmarkLionWithAccumulatedContext(b, newLionJSON())
}

func BenchmarkLionWithAccumulatedContextText(b *testing.B) {
	benchmarkLionWithAccumulatedContext(b, newLionText())
}

func benchmarkLionWithAccumulatedContext(b *testing.B, baseLogger lion.Logger) {
	logger := baseLogger.WithFields(map[string]interface{}{
		"int":               1,
		"int64":             int64(1),
		"float":             3.0,
		"string":            "four!",
		"bool":              true,
		"time":              time.Unix(0, 0),
		"error":             errExample.Error(),
		"duration":          time.Second,
		"user-defined type": _jane,
		"another string":    "done!",
	})
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Infoln("Go really fast.")
		}
	})
}

func BenchmarkLionWithoutFieldsJSON(b *testing.B) {
	benchmarkLionWithoutFields(b, newLionJSON())
}

func BenchmarkLionWithoutFieldsText(b *testing.B) {
	benchmarkLionWithoutFields(b, newLionText())
}

func benchmarkLionWithoutFields(b *testing.B, logger lion.Logger) {
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Infoln("Go fast.")
		}
	})
}

func newLionJSON() lion.Logger {
	return lion.NewLogger(lion.NewJSONWritePusher(ioutil.Discard))
}

func newLionText() lion.Logger {
	return lion.NewLogger(lion.NewTextWritePusher(ioutil.Discard))
}
