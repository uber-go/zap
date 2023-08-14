package lazylogger

import (
	"runtime"
	"strconv"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/internal/ztest"
	"go.uber.org/zap/zapcore"
)

func Benchmark5WithsUsed(b *testing.B) {
	benchmarkWithUsed(b, (*LazyLogger).With, 5, true)
}

func Benchmark5WithsNotUsed(b *testing.B) {
	benchmarkWithUsed(b, (*LazyLogger).With, 5, false)
}

func Benchmark5WithLazysUsed(b *testing.B) {
	benchmarkWithUsed(b, (*LazyLogger).WithLazy, 5, true)
}

func Benchmark5WithLazysNotUsed(b *testing.B) {
	benchmarkWithUsed(b, (*LazyLogger).WithLazy, 5, false)
}

func benchmarkWithUsed(b *testing.B, withMethodExpr func(*LazyLogger, ...zapcore.Field) *LazyLogger, N int, use bool) {
	keys := make([]string, N)
	values := make([]string, N)
	for i := 0; i < N; i++ {
		keys[i] = "k" + strconv.Itoa(i)
		values[i] = "v" + strconv.Itoa(i)
	}

	b.ResetTimer()

	withBenchedLogger(b, func(log *LazyLogger) {
		for i := 0; i < N; i++ {
			log = withMethodExpr(log, zap.String(keys[i], values[i]))
		}
		if use {
			log.Info("used")
			return
		}
		runtime.KeepAlive(log)
	})
}

func withBenchedLogger(b *testing.B, f func(*LazyLogger)) {
	logger := New(
		zapcore.NewCore(
			zapcore.NewJSONEncoder(zap.NewProductionConfig().EncoderConfig),
			&ztest.Discarder{},
			zap.DebugLevel,
		))
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			f(logger)
		}
	})
}
