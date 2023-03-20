package zapslog_test

import (
	"io"
	"testing"

	"go.uber.org/zap/exp/zapslog"
	"go.uber.org/zap/zapcore"
	"golang.org/x/exp/slog"
)

func BenchmarkSlog(b *testing.B) {
	var (
		enc = zapcore.NewJSONEncoder(zapcore.EncoderConfig{
			MessageKey: "msg",
			LevelKey:   "level",
		})
		core   = zapcore.NewCore(enc, zapcore.AddSync(io.Discard), zapcore.DebugLevel)
		logger = slog.New(zapslog.NewHandler(core))
		attrs  = []any{
			slog.String("hello", "world"),
			slog.Group(
				"nested group",
				slog.String("foo", "bar"),
				slog.String("baz", "bat"),
			),
			slog.Group(
				"", // inline group
				slog.String("foo", "bar"),
				slog.String("baz", "bat"),
			),
		}
	)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		logger.Info("hello", attrs...)
	}
}
