package lazylogger

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/internal/ztest"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestLoggerWith(t *testing.T) {
	fieldOpts := opts(zap.Fields(zap.Int("foo", 42)))
	for _, tt := range []struct {
		name           string
		withMethodExpr func(*LazyLogger, ...zap.Field) *LazyLogger
	}{
		{
			"regular non lazy logger",
			(*LazyLogger).With,
		},
		{
			"lazy with logger",
			(*LazyLogger).WithLazy,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			withLogger(t, zap.DebugLevel, fieldOpts, func(logger *LazyLogger, logs *observer.ObservedLogs) {
				// Child loggers should have copy-on-write semantics, so two children
				// shouldn't stomp on each other's fields or affect the parent's fields.
				tt.withMethodExpr(logger, zap.String("one", "two")).Info("")
				tt.withMethodExpr(logger, zap.String("three", "four")).Info("")
				tt.withMethodExpr(logger, zap.String("five", "six")).With(zap.String("seven", "eight")).Info("")
				logger.Info("")

				assert.Equal(t, []observer.LoggedEntry{
					{Context: []zap.Field{zap.Int("foo", 42), zap.String("one", "two")}},
					{Context: []zap.Field{zap.Int("foo", 42), zap.String("three", "four")}},
					{Context: []zap.Field{zap.Int("foo", 42), zap.String("five", "six"), zap.String("seven", "eight")}},
					{Context: []zap.Field{zap.Int("foo", 42)}},
				}, logs.AllUntimed(), "Unexpected cross-talk between child loggers.")
			})
		})
	}
}

func withLogger(t testing.TB, e zapcore.LevelEnabler, opts []zap.Option, f func(*LazyLogger, *observer.ObservedLogs)) {
	fac, logs := observer.New(e)
	log := New(fac, opts...)
	f(log, logs)
}

func opts(opts ...zap.Option) []zap.Option {
	return opts
}

func TestLoggerWithCaptures(t *testing.T) {
	for _, tt := range []struct {
		name           string
		withMethodExpr func(*LazyLogger, ...zap.Field) *LazyLogger
		wantJSON       [2]string
	}{
		{
			name:           "regular with captures arguments at time of With",
			withMethodExpr: (*LazyLogger).With,
			wantJSON: [2]string{
				`{
					"m": "hello",
					"a": [0],
					"b": [1]
				}`,
				`{
					"m": "world",
					"a": [0],
					"c": [2]
				}`,
			},
		},
		{
			name:           "lazy with captures arguments at time of With or Logging",
			withMethodExpr: (*LazyLogger).WithLazy,
			wantJSON: [2]string{
				`{
					"m": "hello",
					"a": [1],
					"b": [1]
				}`,
				`{
					"m": "world",
					"a": [1],
					"c": [2]
				}`,
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			enc := zapcore.NewJSONEncoder(zapcore.EncoderConfig{
				MessageKey: "m",
			})

			var bs ztest.Buffer
			logger := New(zapcore.NewCore(enc, &bs, zap.DebugLevel))

			x := 0
			arr := zapcore.ArrayMarshalerFunc(func(enc zapcore.ArrayEncoder) error {
				enc.AppendInt(x)
				return nil
			})

			// Demonstrate the arguments are captured when With() and Info() are invoked.
			logger = tt.withMethodExpr(logger, zap.Array("a", arr))
			x = 1
			logger.Info("hello", zap.Array("b", arr))
			x = 2
			logger = tt.withMethodExpr(logger, zap.Array("c", arr))
			logger.Info("world")

			if lines := bs.Lines(); assert.Len(t, lines, 2) {
				assert.JSONEq(t, tt.wantJSON[0], lines[0], "Unexpected output from first log.")
				assert.JSONEq(t, tt.wantJSON[1], lines[1], "Unexpected output from second log.")
			}
		})
	}
}
