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
	"errors"
	"testing"
	"time"

	"github.com/uber-go/zap"
)

type user struct {
	Name      string
	Email     string
	CreatedAt time.Time
}

func (u *user) MarshalLog(kv zap.KeyValue) error {
	kv.AddString("name", u.Name)
	kv.AddString("email", u.Email)
	kv.AddInt64("created_at", u.CreatedAt.UnixNano())
	return nil
}

var _jane = &user{
	Name:      "Jane Doe",
	Email:     "jane@test.com",
	CreatedAt: time.Date(1980, 1, 1, 12, 0, 0, 0, time.UTC),
}

func withBenchedLogger(b *testing.B, f func(zap.Logger)) {
	logger := zap.New(
		zap.NewJSONEncoder(),
		zap.DebugLevel,
		zap.DiscardOutput,
	)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			f(logger)
		}
	})
}

func BenchmarkNoContext(b *testing.B) {
	withBenchedLogger(b, func(log zap.Logger) {
		log.Info("No context.")
	})
}

func BenchmarkBoolField(b *testing.B) {
	withBenchedLogger(b, func(log zap.Logger) {
		log.Info("Boolean.", zap.Bool("foo", true))
	})
}

func BenchmarkFloat64Field(b *testing.B) {
	withBenchedLogger(b, func(log zap.Logger) {
		log.Info("Floating point.", zap.Float64("foo", 3.14))
	})
}

func BenchmarkIntField(b *testing.B) {
	withBenchedLogger(b, func(log zap.Logger) {
		log.Info("Integer.", zap.Int("foo", 42))
	})
}

func BenchmarkInt64Field(b *testing.B) {
	withBenchedLogger(b, func(log zap.Logger) {
		log.Info("64-bit integer.", zap.Int64("foo", 42))
	})
}

func BenchmarkStringField(b *testing.B) {
	withBenchedLogger(b, func(log zap.Logger) {
		log.Info("Strings.", zap.String("foo", "bar"))
	})
}

func BenchmarkStringerField(b *testing.B) {
	withBenchedLogger(b, func(log zap.Logger) {
		log.Info("Level.", zap.Stringer("foo", zap.InfoLevel))
	})
}

func BenchmarkTimeField(b *testing.B) {
	t := time.Unix(0, 0)
	withBenchedLogger(b, func(log zap.Logger) {
		log.Info("Time.", zap.Time("foo", t))
	})
}

func BenchmarkDurationField(b *testing.B) {
	withBenchedLogger(b, func(log zap.Logger) {
		log.Info("Duration", zap.Duration("foo", time.Second))
	})
}

func BenchmarkErrorField(b *testing.B) {
	err := errors.New("egad!")
	withBenchedLogger(b, func(log zap.Logger) {
		log.Info("Error.", zap.Error(err))
	})
}

func BenchmarkStackField(b *testing.B) {
	withBenchedLogger(b, func(log zap.Logger) {
		log.Info("Error.", zap.Stack())
	})
}

func BenchmarkMarshalerField(b *testing.B) {
	withBenchedLogger(b, func(log zap.Logger) {
		log.Info("Arbitrary zap.LogMarshaler.", zap.Marshaler("user", _jane))
	})
}

func BenchmarkObjectField(b *testing.B) {
	withBenchedLogger(b, func(log zap.Logger) {
		log.Info("Reflection-based serialization.", zap.Object("user", _jane))
	})
}

func BenchmarkAddCallerHook(b *testing.B) {
	logger := zap.New(
		zap.NewJSONEncoder(),
		zap.DiscardOutput,
		zap.AddCaller(),
	)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info("Caller.")
		}
	})
}

func Benchmark10Fields(b *testing.B) {
	withBenchedLogger(b, func(log zap.Logger) {
		log.Info("Ten fields, passed at the log site.",
			zap.Int("one", 1),
			zap.Int("two", 2),
			zap.Int("three", 3),
			zap.Int("four", 4),
			zap.Int("five", 5),
			zap.Int("six", 6),
			zap.Int("seven", 7),
			zap.Int("eight", 8),
			zap.Int("nine", 9),
			zap.Int("ten", 10),
		)
	})
}

func Benchmark100Fields(b *testing.B) {
	const batchSize = 50
	logger := zap.New(
		zap.NewJSONEncoder(),
		zap.DebugLevel,
		zap.DiscardOutput,
	)

	// Don't include allocating these helper slices in the benchmark. Since
	// access to them isn't synchronized, we can't run the benchmark in
	// parallel.
	first := make([]zap.Field, batchSize)
	second := make([]zap.Field, batchSize)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for i := 0; i < batchSize; i++ {
			// We're duplicating keys, but that doesn't affect performance.
			first[i] = zap.Int("foo", i)
			second[i] = zap.Int("foo", i+batchSize)
		}
		logger.With(first...).Info("Child loggers with lots of context.", second...)
	}
}
