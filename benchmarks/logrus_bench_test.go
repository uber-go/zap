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

	"github.com/uber-go/zap"
	"github.com/uber-go/zap/zbark"

	"github.com/Sirupsen/logrus"
	"github.com/uber-common/bark"
)

func newLogrus() *logrus.Logger {
	return &logrus.Logger{
		Out:       ioutil.Discard,
		Formatter: new(logrus.JSONFormatter),
		Hooks:     make(logrus.LevelHooks),
		Level:     logrus.DebugLevel,
	}
}

func BenchmarkLogrusAddingFields(b *testing.B) {
	logger := newLogrus()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.WithFields(logrus.Fields{
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
			}).Info("Go fast.")
		}
	})
}

func BenchmarkZapBarkifyAddingFields(b *testing.B) {
	logger := zbark.Barkify(zap.NewJSON(zap.DebugLevel, zap.Output(zap.Discard)))
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.WithFields(bark.Fields{
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
			}).Info("Go fast.")
		}
	})
}

func BenchmarkLogrusWithAccumulatedContext(b *testing.B) {
	baseLogger := newLogrus()
	logger := baseLogger.WithFields(logrus.Fields{
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
			logger.Info("Go really fast.")
		}
	})
}

func BenchmarkZapBarkifyWithAccumulatedContext(b *testing.B) {
	baseLogger := zbark.Barkify(zap.NewJSON(zap.DebugLevel, zap.Output(zap.Discard)))
	logger := baseLogger.WithFields(bark.Fields{
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
			logger.Info("Go really fast.")
		}
	})
}

func BenchmarkLogrusWithoutFields(b *testing.B) {
	logger := newLogrus()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info("Go fast.")
		}
	})
}
