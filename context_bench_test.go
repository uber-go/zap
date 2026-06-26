// Copyright (c) 2026 Uber Technologies, Inc.
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

	"go.uber.org/zap/internal/ztest"
	"go.uber.org/zap/zapcore"
)

// benchedContextLogger builds a Logger that discards output, optionally with
// the supplied options, for benchmarking the *Ctx code paths in isolation.
func benchedContextLogger(opts ...Option) *Logger {
	return New(
		zapcore.NewCore(
			zapcore.NewJSONEncoder(NewProductionConfig().EncoderConfig),
			&ztest.Discarder{},
			DebugLevel,
		),
		opts...,
	)
}

// runParallel runs f under the standard parallel benchmark harness.
func runParallel(b *testing.B, f func()) {
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			f()
		}
	})
}

// The benchmarks below compare, for each logging style, three configurations:
//
//   - Baseline:       the existing non-context method.
//   - Ctx_NoOption:   the *Ctx method on a logger with NO Context option set,
//                     which measures the overhead of the nil-check fast path.
//   - Ctx_WithOption: the *Ctx method on a logger configured with a Context
//                     option that extracts one field, which measures the full
//                     extraction + prepend cost.
//
// All log levels (Debug/Info/Warn/Error/...) share an identical code path that
// only differs by the level constant, so Info is used as the representative
// level for every style.

func BenchmarkLoggerContext(b *testing.B) {
	ctx := ctxWithCorrelationID()
	field := String("extra", "value")

	b.Run("Info/Baseline", func(b *testing.B) {
		log := benchedContextLogger()
		runParallel(b, func() { log.Info("msg", field) })
	})
	b.Run("InfoCtx/NoOption", func(b *testing.B) {
		log := benchedContextLogger()
		runParallel(b, func() { log.InfoCtx(ctx, "msg", field) })
	})
	b.Run("InfoCtx/WithOption", func(b *testing.B) {
		log := benchedContextLogger(correlationOption())
		runParallel(b, func() { log.InfoCtx(ctx, "msg", field) })
	})

	b.Run("Log/Baseline", func(b *testing.B) {
		log := benchedContextLogger()
		runParallel(b, func() { log.Log(InfoLevel, "msg", field) })
	})
	b.Run("LogCtx/WithOption", func(b *testing.B) {
		log := benchedContextLogger(correlationOption())
		runParallel(b, func() { log.LogCtx(ctx, InfoLevel, "msg", field) })
	})
}

func BenchmarkSugaredContext(b *testing.B) {
	ctx := ctxWithCorrelationID()

	// print-style
	b.Run("Print/Baseline", func(b *testing.B) {
		s := benchedContextLogger().Sugar()
		runParallel(b, func() { s.Info("msg") })
	})
	b.Run("PrintCtx/NoOption", func(b *testing.B) {
		s := benchedContextLogger().Sugar()
		runParallel(b, func() { s.InfoCtx(ctx, "msg") })
	})
	b.Run("PrintCtx/WithOption", func(b *testing.B) {
		s := benchedContextLogger(correlationOption()).Sugar()
		runParallel(b, func() { s.InfoCtx(ctx, "msg") })
	})

	// w-style
	b.Run("Infow/Baseline", func(b *testing.B) {
		s := benchedContextLogger().Sugar()
		runParallel(b, func() { s.Infow("msg", "extra", "value") })
	})
	b.Run("InfowCtx/NoOption", func(b *testing.B) {
		s := benchedContextLogger().Sugar()
		runParallel(b, func() { s.InfowCtx(ctx, "msg", "extra", "value") })
	})
	b.Run("InfowCtx/WithOption", func(b *testing.B) {
		s := benchedContextLogger(correlationOption()).Sugar()
		runParallel(b, func() { s.InfowCtx(ctx, "msg", "extra", "value") })
	})

	// f-style
	b.Run("Infof/Baseline", func(b *testing.B) {
		s := benchedContextLogger().Sugar()
		runParallel(b, func() { s.Infof("msg %d", 1) })
	})
	b.Run("InfofCtx/WithOption", func(b *testing.B) {
		s := benchedContextLogger(correlationOption()).Sugar()
		runParallel(b, func() { s.InfofCtx(ctx, "msg %d", 1) })
	})

	// ln-style
	b.Run("Infoln/Baseline", func(b *testing.B) {
		s := benchedContextLogger().Sugar()
		runParallel(b, func() { s.Infoln("msg") })
	})
	b.Run("InfolnCtx/WithOption", func(b *testing.B) {
		s := benchedContextLogger(correlationOption()).Sugar()
		runParallel(b, func() { s.InfolnCtx(ctx, "msg") })
	})
}
