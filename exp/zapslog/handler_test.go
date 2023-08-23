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
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestAddCaller(t *testing.T) {
	fac, logs := observer.New(zapcore.DebugLevel)
	sl := slog.New(NewHandler(fac, AddCaller()))
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
	r := require.New(t)
	fac, logs := observer.New(zapcore.DebugLevel)
	sl := slog.New(NewHandler(fac, AddStacktrace(zapcore.DebugLevel)))
	sl.Info("msg")

	r.Len(logs.AllUntimed(), 1, "Expected exactly one entry to be logged")
	entry := logs.AllUntimed()[0]
	r.Equal("msg", entry.Message, "Unexpected message")
	assert.Regexp(t,
		`^go.uber.org/zap/exp/zapslog.TestAddStack`,
		entry.Stack,
		"Unexpected stack trace annotation.",
	)
	assert.Regexp(t,
		`/zapslog/slog_go121_test.go:\d+`,
		entry.Stack,
		"Unexpected stack trace annotation.",
	)
}
