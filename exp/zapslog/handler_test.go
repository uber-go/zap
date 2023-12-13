// Copyright (c) 2023 Uber Technologies, Inc.
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

//go:build go1.21

package zapslog

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

type testContextKey string

func TestAddCaller(t *testing.T) {
	t.Parallel()

	fac, logs := observer.New(zapcore.DebugLevel)
	sl := slog.New(NewHandler(fac, WithCaller(true)))
	sl.Info("msg")

	require.Len(t, logs.AllUntimed(), 1, "Expected exactly one entry to be logged")
	entry := logs.AllUntimed()[0]
	assert.Equal(t, "msg", entry.Message, "Unexpected message")
	assert.Regexp(t,
		`/handler_test.go:\d+$`,
		entry.Caller.String(),
		"Unexpected caller annotation.",
	)
}

func TestAddStack(t *testing.T) {
	fac, logs := observer.New(zapcore.DebugLevel)
	sl := slog.New(NewHandler(fac, AddStacktraceAt(slog.LevelDebug)))
	sl.Info("msg")

	require.Len(t, logs.AllUntimed(), 1, "Expected exactly one entry to be logged")
	entry := logs.AllUntimed()[0]
	require.Equal(t, "msg", entry.Message, "Unexpected message")
	assert.Regexp(t,
		`^go.uber.org/zap/exp/zapslog.TestAddStack`,
		entry.Stack,
		"Unexpected stack trace annotation.",
	)
	assert.Regexp(t,
		`/zapslog/handler_test.go:\d+`,
		entry.Stack,
		"Unexpected stack trace annotation.",
	)
}

func TestAddStackSkip(t *testing.T) {
	fac, logs := observer.New(zapcore.DebugLevel)
	sl := slog.New(NewHandler(fac, AddStacktraceAt(slog.LevelDebug), WithCallerSkip(1)))
	sl.Info("msg")

	require.Len(t, logs.AllUntimed(), 1, "Expected exactly one entry to be logged")
	entry := logs.AllUntimed()[0]
	assert.Regexp(t,
		`src/testing/testing.go:\d+`,
		entry.Stack,
		"Unexpected stack trace annotation.",
	)
}

func TestEmptyAttr(t *testing.T) {
	t.Parallel()

	fac, observedLogs := observer.New(zapcore.DebugLevel)
	sl := slog.New(NewHandler(fac))

	t.Run("Handle", func(t *testing.T) {
		sl.Info(
			"msg",
			slog.String("foo", "bar"),
			slog.Attr{},
		)

		logs := observedLogs.TakeAll()
		require.Len(t, logs, 1, "Expected exactly one entry to be logged")
		assert.Equal(t, map[string]any{
			"foo": "bar",
		}, logs[0].ContextMap(), "Unexpected context")
	})

	t.Run("WithAttrs", func(t *testing.T) {
		sl.With(slog.String("foo", "bar"), slog.Attr{}).Info("msg")

		logs := observedLogs.TakeAll()
		require.Len(t, logs, 1, "Expected exactly one entry to be logged")
		assert.Equal(t, map[string]any{
			"foo": "bar",
		}, logs[0].ContextMap(), "Unexpected context")
	})

	t.Run("Group", func(t *testing.T) {
		sl.With("k", slog.GroupValue(slog.String("foo", "bar"), slog.Attr{})).Info("msg")

		logs := observedLogs.TakeAll()
		require.Len(t, logs, 1, "Expected exactly one entry to be logged")
		assert.Equal(t, map[string]any{
			"k": map[string]any{
				"foo": "bar",
			},
		}, logs[0].ContextMap(), "Unexpected context")
	})
}

func TestWithName(t *testing.T) {
	t.Parallel()
	fac, observedLogs := observer.New(zapcore.DebugLevel)
	t.Run("default", func(t *testing.T) {
		sl := slog.New(NewHandler(fac))
		sl.Info("msg")

		logs := observedLogs.TakeAll()
		require.Len(t, logs, 1, "Expected exactly one entry to be logged")
		entry := logs[0]
		assert.Equal(t, "", entry.LoggerName, "Unexpected logger name")
	})
	t.Run("with name", func(t *testing.T) {
		sl := slog.New(NewHandler(fac, WithName("test-name")))
		sl.Info("msg")

		logs := observedLogs.TakeAll()
		require.Len(t, logs, 1, "Expected exactly one entry to be logged")
		entry := logs[0]
		assert.Equal(t, "test-name", entry.LoggerName, "Unexpected logger name")
	})
}

type Token string

func (Token) LogValue() slog.Value {
	return slog.StringValue("REDACTED_TOKEN")
}

func TestAttrKinds(t *testing.T) {
	fac, logs := observer.New(zapcore.DebugLevel)
	sl := slog.New(NewHandler(fac))
	testToken := Token("no")
	sl.Info(
		"msg",
		slog.Bool("bool", true),
		slog.Duration("duration", time.Hour),
		slog.Float64("float64", 42.0),
		slog.Int64("int64", -1234),
		slog.Time("time", time.Date(2015, 10, 21, 7, 28, 0o0, 0, time.UTC)),
		slog.Uint64("uint64", 2),
		slog.Group("group", slog.String("inner", "inner-group")),
		"logvaluer", testToken,
		"any", "what am i?",
	)

	require.Len(t, logs.AllUntimed(), 1, "Expected exactly one entry to be logged")
	entry := logs.AllUntimed()[0]
	assert.Equal(t,
		map[string]any{
			"bool":      true,
			"duration":  time.Hour,
			"float64":   float64(42),
			"group":     map[string]any{"inner": "inner-group"},
			"int64":     int64(-1234),
			"time":      time.Date(2015, time.October, 21, 7, 28, 0, 0, time.UTC),
			"uint64":    uint64(2),
			"logvaluer": "REDACTED_TOKEN",
			"any":       "what am i?",
		},
		entry.ContextMap())
}

func TestContextFieldExtractor(t *testing.T) {
	key := testContextKey("testkey")
	fac, logs := observer.New(zapcore.DebugLevel)
	ctx := context.WithValue(context.Background(), key, "testvalue")

	sl := slog.New(NewHandler(fac, WithContextFieldExtractors(func(ctx context.Context) []zapcore.Field {
		v := ctx.Value(key).(string)
		return []zapcore.Field{zap.String("testkey", v)}
	})))
	sl.InfoContext(ctx, "msg")
	lines := logs.TakeAll()

	require.Len(t, lines, 1)
	require.Equal(t, "testvalue", lines[0].ContextMap()["testkey"])
}
