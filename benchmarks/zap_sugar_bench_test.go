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

/* FIXME (shah): Update to match new APIs.

import (
	"testing"
	"time"

	"github.com/uber-go/zap"
	"github.com/uber-go/zap/zwrap"
)

func fakeSugarFields() []interface{} {
	return []interface{}{
		errExample,
		"int", 1,
		"int64", 2,
		"float", 3.0,
		"string", "four!",
		"stringer", zap.DebugLevel,
		"bool", true,
		"time", time.Unix(0, 0),
		"duration", time.Second,
		"another string", "done!",
	}
}

var newCoreLoggerDefaults = []zap.Option{
	zap.ErrorLevel,
	zap.DiscardOutput,
}

func newCoreLogger(options ...zap.Option) zap.Logger {
	options = append(newCoreLoggerDefaults, options...)
	return zap.New(zap.NewJSONEncoder(), options...)
}

func newSugarLogger(options ...zap.Option) zap.Sugar {
	return zap.NewSugar(newCoreLogger(options...))
}

func BenchmarkZapSugarDisabledLevelsWithoutFields(b *testing.B) {
	logger := newSugarLogger()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info("Should be discarded.")
		}
	})
}

func BenchmarkZapSugarDisabledLevelsAccumulatedContext(b *testing.B) {
	context := fakeFields()
	logger := newSugarLogger(zap.Fields(context...))
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info("Should be discarded.")
		}
	})
}

func BenchmarkZapSugarDisabledLevelsAddingFields(b *testing.B) {
	logger := newSugarLogger()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info("Should be discarded.", fakeSugarFields()...)
		}
	})
}

func BenchmarkZapSugarAddingFields(b *testing.B) {
	logger := newSugarLogger()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info("Go fast.", fakeSugarFields()...)
		}
	})
}

func BenchmarkZapSugarWithAccumulatedContext(b *testing.B) {
	logger := newSugarLogger().With(fakeSugarFields()...)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info("Go really fast.")
		}
	})
}

func BenchmarkZapSugarWithoutFields(b *testing.B) {
	logger := newSugarLogger()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info("Go fast.")
		}
	})
}

func BenchmarkZapSugarSampleWithoutFields(b *testing.B) {
	messages := fakeMessages(1000)
	core := newCoreLogger()
	logger := zap.NewSugar(zwrap.Sample(core, time.Second, 10, 10000))
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			i++
			logger.Info(messages[i%1000])
		}
	})
}

func BenchmarkZapSugarSampleAddingFields(b *testing.B) {
	messages := fakeMessages(1000)
	core := newCoreLogger()
	logger := zap.NewSugar(zwrap.Sample(core, time.Second, 10, 10000))
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			i++
			logger.Info(messages[i%1000], fakeSugarFields()...)
		}
	})
}
*/
