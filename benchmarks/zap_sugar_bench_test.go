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
	"testing"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/testutils"
	"go.uber.org/zap/zapcore"
)

func fakeSugarFields() []interface{} {
	return []interface{}{
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

func fakeSugarFormatArgs() (string, []interface{}) {
	template := "some args: %d %v %v %s %v %v %v %v %v"
	args := []interface{}{1, 2, 3.0, "four!", zap.DebugLevel, true, time.Unix(0, 0), time.Second, "done!"}
	return template, args
}

func newSugarLogger(lvl zapcore.Level, options ...zap.Option) *zap.SugaredLogger {
	return zap.New(zapcore.WriterFacility(
		benchEncoder(),
		&testutils.Discarder{},
		lvl,
	), options...).Sugar()
}

func BenchmarkZapSugarDisabledLevelsWithoutFields(b *testing.B) {
	logger := newSugarLogger(zap.ErrorLevel)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info("Should be discarded.")
		}
	})
}

func BenchmarkZapSugarFmtDisabledLevelsWithoutFields(b *testing.B) {
	logger := newSugarLogger(zap.ErrorLevel)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			template, args := fakeSugarFormatArgs()
			logger.Infof(template, args...)
		}
	})
}

func BenchmarkZapSugarDisabledLevelsAccumulatedContext(b *testing.B) {
	context := fakeFields()
	logger := newSugarLogger(zap.ErrorLevel, zap.Fields(context...))
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info("Should be discarded.")
		}
	})
}

func BenchmarkZapSugarFmtDisabledLevelsAccumulatedContext(b *testing.B) {
	context := fakeFields()
	logger := newSugarLogger(zap.ErrorLevel, zap.Fields(context...))
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			template, args := fakeSugarFormatArgs()
			logger.Infof(template, args...)
		}
	})
}

func BenchmarkZapSugarDisabledLevelsAddingFields(b *testing.B) {
	logger := newSugarLogger(zap.ErrorLevel)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Infow("Should be discarded.", fakeSugarFields()...)
		}
	})
}

func BenchmarkZapSugarAddingFields(b *testing.B) {
	logger := newSugarLogger(zap.DebugLevel)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Infow("Go fast.", fakeSugarFields()...)
		}
	})
}

func BenchmarkZapSugarWithAccumulatedContext(b *testing.B) {
	logger := newSugarLogger(zap.DebugLevel).With(fakeSugarFields()...)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info("Go really fast.")
		}
	})
}

func BenchmarkZapSugarFmtWithAccumulatedContext(b *testing.B) {
	logger := newSugarLogger(zap.DebugLevel).With(fakeSugarFields()...)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			template, args := fakeSugarFormatArgs()
			logger.Infof(template, args...)
		}
	})
}

func BenchmarkZapSugarWithoutFields(b *testing.B) {
	logger := newSugarLogger(zap.DebugLevel)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info("Go fast.")
		}
	})
}

func BenchmarkZapSugarFmtWithoutFields(b *testing.B) {
	logger := newSugarLogger(zap.DebugLevel)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			template, args := fakeSugarFormatArgs()
			logger.Infof(template, args...)
		}
	})
}
