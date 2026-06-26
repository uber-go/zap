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
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/internal/exit"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

type ctxKey string

const correlationIDKey ctxKey = "correlation-id"

const correlationID = "e9718aab-aa1a-4ad8-b1e6-690f6c43bd15"

// correlationOption returns a Context option that promotes the correlation ID
// stored on the context (if any) to a structured field.
func correlationOption() Option {
	return Context(func(ctx context.Context) []Field {
		if id, ok := ctx.Value(correlationIDKey).(string); ok {
			return []Field{String(string(correlationIDKey), id)}
		}
		return nil
	})
}

func ctxWithCorrelationID() context.Context {
	return context.WithValue(context.Background(), correlationIDKey, correlationID)
}

// fieldMap collapses an observed entry's context into a key->value map for
// convenient assertions.
func fieldMap(fields []Field) map[string]interface{} {
	enc := zapcore.NewMapObjectEncoder()
	for _, f := range fields {
		f.AddTo(enc)
	}
	return enc.Fields
}

func TestLoggerContextFields(t *testing.T) {
	t.Run("context fields are prepended to log-site fields", func(t *testing.T) {
		core, logs := observer.New(zapcore.DebugLevel)
		logger := New(core, correlationOption())
		ctx := ctxWithCorrelationID()

		logger.InfoCtx(ctx, "hello", String("extra", "value"))

		entries := logs.All()
		require.Len(t, entries, 1)
		// Context fields come first, then log-site fields.
		require.Len(t, entries[0].Context, 2)
		assert.Equal(t, string(correlationIDKey), entries[0].Context[0].Key)
		assert.Equal(t, "extra", entries[0].Context[1].Key)

		got := fieldMap(entries[0].Context)
		assert.Equal(t, correlationID, got[string(correlationIDKey)])
		assert.Equal(t, "value", got["extra"])
	})

	t.Run("every leveled Ctx method emits context fields", func(t *testing.T) {
		core, logs := observer.New(zapcore.DebugLevel)
		logger := New(core, correlationOption())
		ctx := ctxWithCorrelationID()

		logger.DebugCtx(ctx, "m")
		logger.InfoCtx(ctx, "m")
		logger.WarnCtx(ctx, "m")
		logger.ErrorCtx(ctx, "m")
		logger.LogCtx(ctx, InfoLevel, "m")

		entries := logs.All()
		require.Len(t, entries, 5)
		for _, e := range entries {
			assert.Equal(t, correlationID, fieldMap(e.Context)[string(correlationIDKey)])
		}
	})

	// Regression test for the latent bug in the original PR #1019: when no
	// Context option is configured (or ctx is nil), the explicit log-site
	// fields must still be emitted instead of being dropped.
	t.Run("log-site fields preserved without Context option", func(t *testing.T) {
		core, logs := observer.New(zapcore.DebugLevel)
		logger := New(core) // no Context option
		ctx := ctxWithCorrelationID()

		logger.InfoCtx(ctx, "hello", String("extra", "value"))

		entries := logs.All()
		require.Len(t, entries, 1)
		require.Len(t, entries[0].Context, 1)
		assert.Equal(t, "value", fieldMap(entries[0].Context)["extra"])
	})

	t.Run("log-site fields preserved with nil context", func(t *testing.T) {
		core, logs := observer.New(zapcore.DebugLevel)
		logger := New(core, correlationOption())

		logger.InfoCtx(nil, "hello", String("extra", "value")) //nolint:staticcheck // intentionally nil

		entries := logs.All()
		require.Len(t, entries, 1)
		require.Len(t, entries[0].Context, 1)
		assert.Equal(t, "value", fieldMap(entries[0].Context)["extra"])
	})

	t.Run("DPanic/Panic/Fatal Ctx methods", func(t *testing.T) {
		core, logs := observer.New(zapcore.DebugLevel)
		logger := New(core, correlationOption())
		ctx := ctxWithCorrelationID()

		logger.DPanicCtx(ctx, "dpanic")
		assert.Panics(t, func() { logger.PanicCtx(ctx, "panic") })
		stub := exit.WithStub(func() { logger.FatalCtx(ctx, "fatal") })
		assert.True(t, stub.Exited, "Expected FatalCtx to exit the process.")

		for _, e := range logs.All() {
			assert.Equal(t, correlationID, fieldMap(e.Context)[string(correlationIDKey)])
		}
	})
}

func TestSugaredLoggerContextFields(t *testing.T) {
	ctx := ctxWithCorrelationID()

	t.Run("w-style methods prepend context fields", func(t *testing.T) {
		core, logs := observer.New(zapcore.DebugLevel)
		sugar := New(core, correlationOption()).Sugar()

		sugar.InfowCtx(ctx, "hello", "extra", "value")

		entries := logs.All()
		require.Len(t, entries, 1)
		got := fieldMap(entries[0].Context)
		assert.Equal(t, correlationID, got[string(correlationIDKey)])
		assert.Equal(t, "value", got["extra"])
	})

	t.Run("f-style methods attach context fields", func(t *testing.T) {
		core, logs := observer.New(zapcore.DebugLevel)
		sugar := New(core, correlationOption()).Sugar()

		sugar.InfofCtx(ctx, "hello %s", "world")

		entries := logs.All()
		require.Len(t, entries, 1)
		assert.Equal(t, "hello world", entries[0].Message)
		assert.Equal(t, correlationID, fieldMap(entries[0].Context)[string(correlationIDKey)])
	})

	t.Run("print-style and ln-style methods attach context fields", func(t *testing.T) {
		core, logs := observer.New(zapcore.DebugLevel)
		sugar := New(core, correlationOption()).Sugar()

		sugar.InfoCtx(ctx, "print style")
		sugar.InfolnCtx(ctx, "ln style")

		entries := logs.All()
		require.Len(t, entries, 2)
		for _, e := range entries {
			assert.Equal(t, correlationID, fieldMap(e.Context)[string(correlationIDKey)])
		}
	})

	t.Run("Log/Logw/Logf/Logln Ctx variants", func(t *testing.T) {
		core, logs := observer.New(zapcore.DebugLevel)
		sugar := New(core, correlationOption()).Sugar()

		sugar.LogCtx(ctx, InfoLevel, "m")
		sugar.LogwCtx(ctx, InfoLevel, "m", "extra", "value")
		sugar.LogfCtx(ctx, InfoLevel, "m %d", 1)
		sugar.LoglnCtx(ctx, InfoLevel, "m")

		entries := logs.All()
		require.Len(t, entries, 4)
		for _, e := range entries {
			assert.Equal(t, correlationID, fieldMap(e.Context)[string(correlationIDKey)])
		}
	})

	t.Run("w-style log-site pairs preserved without Context option", func(t *testing.T) {
		core, logs := observer.New(zapcore.DebugLevel)
		sugar := New(core).Sugar() // no Context option

		sugar.InfowCtx(ctx, "hello", "extra", "value")

		entries := logs.All()
		require.Len(t, entries, 1)
		require.Len(t, entries[0].Context, 1)
		assert.Equal(t, "value", fieldMap(entries[0].Context)["extra"])
	})

	t.Run("Panic/Fatal w-style Ctx methods", func(t *testing.T) {
		core, logs := observer.New(zapcore.DebugLevel)
		sugar := New(core, correlationOption()).Sugar()

		assert.Panics(t, func() { sugar.PanicwCtx(ctx, "panic", "extra", "value") })
		stub := exit.WithStub(func() { sugar.FatalwCtx(ctx, "fatal") })
		assert.True(t, stub.Exited, "Expected FatalwCtx to exit the process.")

		for _, e := range logs.All() {
			assert.Equal(t, correlationID, fieldMap(e.Context)[string(correlationIDKey)])
		}
	})
}
