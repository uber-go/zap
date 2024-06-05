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

package zapslog_test

import (
	"context"
	"log/slog"
	"net"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/exp/zapslog"
)

type Password string

func (p Password) LogValue() slog.Value {
	return slog.StringValue("REDACTED")
}

func Example_slog() {
	logger := zap.NewExample(zap.IncreaseLevel(zap.InfoLevel))
	defer logger.Sync()

	sl := slog.New(zapslog.NewHandler(logger.Core()))
	ctx := context.Background()

	sl.Info("user", "name", "Al", "secret", Password("secret"))
	sl.Error("oops", "err", net.ErrClosed, "status", 500)
	sl.LogAttrs(
		ctx,
		slog.LevelError,
		"oops",
		slog.Any("err", net.ErrClosed),
		slog.Int("status", 500),
	)
	sl.Info("message",
		slog.Group("group",
			slog.Float64("pi", 3.14),
			slog.Duration("1min", time.Minute),
		),
	)
	sl.WithGroup("s").LogAttrs(
		ctx,
		slog.LevelWarn,
		"warn msg", // message
		slog.Uint64("u", 1),
		slog.Any("m", map[string]any{
			"foo": "bar",
		}))
	sl.LogAttrs(ctx, slog.LevelDebug, "not show up")

	// Output:
	// {"level":"info","msg":"user","name":"Al","secret":"REDACTED"}
	// {"level":"error","msg":"oops","err":"use of closed network connection","status":500}
	// {"level":"error","msg":"oops","err":"use of closed network connection","status":500}
	// {"level":"info","msg":"message","group":{"pi":3.14,"1min":"1m0s"}}
	// {"level":"warn","msg":"warn msg","s":{"u":1,"m":{"foo":"bar"}}}
}
