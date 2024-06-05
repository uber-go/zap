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
	"bytes"
	"encoding/json"
	"log/slog"
	"sync"
	"testing"
	"testing/slogtest"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"
	"go.uber.org/zap/zaptest/observer"
)

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

func TestInlineGroup(t *testing.T) {
	fac, observedLogs := observer.New(zapcore.DebugLevel)

	t.Run("simple", func(t *testing.T) {
		sl := slog.New(NewHandler(fac))
		sl.Info("msg", "a", "b", slog.Group("", slog.String("c", "d")), "e", "f")

		logs := observedLogs.TakeAll()
		require.Len(t, logs, 1, "Expected exactly one entry to be logged")
		assert.Equal(t, map[string]any{
			"a": "b",
			"c": "d",
			"e": "f",
		}, logs[0].ContextMap(), "Unexpected context")
	})

	t.Run("recursive", func(t *testing.T) {
		sl := slog.New(NewHandler(fac))
		sl.Info("msg", "a", "b", slog.Group("", slog.Group("", slog.Group("", slog.String("c", "d"))), slog.Group("", "e", "f")))

		logs := observedLogs.TakeAll()
		require.Len(t, logs, 1, "Expected exactly one entry to be logged")
		assert.Equal(t, map[string]any{
			"a": "b",
			"c": "d",
			"e": "f",
		}, logs[0].ContextMap(), "Unexpected context")
	})
}

func TestWithGroup(t *testing.T) {
	fac, observedLogs := observer.New(zapcore.DebugLevel)

	// Groups can be nested inside each other.
	t.Run("nested", func(t *testing.T) {
		sl := slog.New(NewHandler(fac))
		sl.With("a", "b").WithGroup("G").WithGroup("in").Info("msg", "c", "d")

		logs := observedLogs.TakeAll()
		require.Len(t, logs, 1, "Expected exactly one entry to be logged")
		assert.Equal(t, map[string]any{
			"G": map[string]any{
				"in": map[string]any{
					"c": "d",
				},
			},
			"a": "b",
		}, logs[0].ContextMap(), "Unexpected context")
	})

	t.Run("nested empty", func(t *testing.T) {
		sl := slog.New(NewHandler(fac))
		sl.With("a", "b").WithGroup("G").WithGroup("in").Info("msg")

		logs := observedLogs.TakeAll()
		require.Len(t, logs, 1, "Expected exactly one entry to be logged")
		assert.Equal(t, map[string]any{
			"a": "b",
		}, logs[0].ContextMap(), "Unexpected context")
	})

	t.Run("empty group", func(t *testing.T) {
		sl := slog.New(NewHandler(fac))
		sl.With("a", "b").WithGroup("G").With("c", "d").WithGroup("H").Info("msg")

		logs := observedLogs.TakeAll()
		require.Len(t, logs, 1, "Expected exactly one entry to be logged")
		assert.Equal(t, map[string]any{
			"G": map[string]any{
				"c": "d",
			},
			"a": "b",
		}, logs[0].ContextMap(), "Unexpected context")
	})

	t.Run("skipped field", func(t *testing.T) {
		sl := slog.New(NewHandler(fac))
		sl.WithGroup("H").With(slog.Attr{}).Info("msg")

		logs := observedLogs.TakeAll()
		require.Len(t, logs, 1, "Expected exactly one entry to be logged")
		assert.Equal(t, map[string]any{}, logs[0].ContextMap(), "Unexpected context")
	})

	t.Run("reuse", func(t *testing.T) {
		sl := slog.New(NewHandler(fac)).WithGroup("G")

		sl.With("a", "b").Info("msg1", "c", "d")
		sl.With("e", "f").Info("msg2", "g", "h")

		logs := observedLogs.TakeAll()
		require.Len(t, logs, 2, "Expected exactly two entries to be logged")

		assert.Equal(t, map[string]any{
			"G": map[string]any{
				"a": "b",
				"c": "d",
			},
		}, logs[0].ContextMap(), "Unexpected context")
		assert.Equal(t, "msg1", logs[0].Message, "Unexpected message")

		assert.Equal(t, map[string]any{
			"G": map[string]any{
				"e": "f",
				"g": "h",
			},
		}, logs[1].ContextMap(), "Unexpected context")
		assert.Equal(t, "msg2", logs[1].Message, "Unexpected message")
	})
}

// Run a few different loggers with concurrent logs
// in an attempt to trip up 'go test -race' and discover any data races.
func TestConcurrentLogs(t *testing.T) {
	t.Parallel()

	const (
		NumWorkers = 10
		NumLogs    = 100
	)

	tests := []struct {
		name         string
		buildHandler func(zapcore.Core) slog.Handler
	}{
		{
			name: "default",
			buildHandler: func(core zapcore.Core) slog.Handler {
				return NewHandler(core)
			},
		},
		{
			name: "grouped",
			buildHandler: func(core zapcore.Core) slog.Handler {
				return NewHandler(core).WithGroup("G")
			},
		},
		{
			name: "named",
			buildHandler: func(core zapcore.Core) slog.Handler {
				return NewHandler(core, WithName("test-name"))
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			fac, observedLogs := observer.New(zapcore.DebugLevel)
			sl := slog.New(tt.buildHandler(fac))

			// Use two wait groups to coordinate the workers:
			//
			// - ready: indicates when all workers should start logging.
			// - done: indicates when all workers have finished logging.
			var ready, done sync.WaitGroup
			ready.Add(NumWorkers)
			done.Add(NumWorkers)

			for i := 0; i < NumWorkers; i++ {
				i := i
				go func() {
					defer done.Done()

					ready.Done() // I'm ready.
					ready.Wait() // Are others?

					for j := 0; j < NumLogs; j++ {
						sl.Info("msg", "worker", i, "log", j)
					}
				}()
			}

			done.Wait()

			// Ensure that all logs were recorded.
			logs := observedLogs.TakeAll()
			assert.Len(t, logs, NumWorkers*NumLogs,
				"Wrong number of logs recorded")
		})
	}
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

func TestSlogtest(t *testing.T) {
	var buff bytes.Buffer
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zapcore.EncoderConfig{
			TimeKey:     slog.TimeKey,
			MessageKey:  slog.MessageKey,
			LevelKey:    slog.LevelKey,
			EncodeLevel: zapcore.CapitalLevelEncoder,
			EncodeTime:  zapcore.RFC3339TimeEncoder,
		}),
		zapcore.AddSync(&buff),
		zapcore.DebugLevel,
	)

	// zaptest doesn't expose the underlying core,
	// so we'll extract it from the logger.
	testCore := zaptest.NewLogger(t).Core()

	handler := NewHandler(zapcore.NewTee(core, testCore))
	err := slogtest.TestHandler(
		handler,
		func() []map[string]any {
			// Parse the newline-delimted JSON in buff.
			var entries []map[string]any

			dec := json.NewDecoder(bytes.NewReader(buff.Bytes()))
			for dec.More() {
				var ent map[string]any
				require.NoError(t, dec.Decode(&ent), "Error decoding log message")
				entries = append(entries, ent)
			}

			return entries
		},
	)
	require.NoError(t, err, "Unexpected error from slogtest.TestHandler")
}
