// Copyright (c) 2017 Uber Technologies, Inc.
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

package zaptest

import (
	"bytes"
	"fmt"
	"strings"
	"sync/atomic"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// LoggerOption configures the test logger built by NewLogger.
type LoggerOption interface {
	applyLoggerOption(*loggerOptions)
}

type loggerOptions struct {
	Level                   zapcore.LevelEnabler
	zapOptions              []zap.Option
	muteAfterTestCompletion bool
}

type loggerOptionFunc func(*loggerOptions)

func (f loggerOptionFunc) applyLoggerOption(opts *loggerOptions) {
	f(opts)
}

// Level controls which messages are logged by a test Logger built by
// NewLogger.
func Level(enab zapcore.LevelEnabler) LoggerOption {
	return loggerOptionFunc(func(opts *loggerOptions) {
		opts.Level = enab
	})
}

// WrapOptions adds zap.Option's to a test Logger built by NewLogger.
func WrapOptions(zapOpts ...zap.Option) LoggerOption {
	return loggerOptionFunc(func(opts *loggerOptions) {
		opts.zapOptions = zapOpts
	})
}

// MuteAfterTestCompletion returns a LoggerOption that suppresses panics
// caused by logging after the test has finished.
//
// By default, testing.TB panics when a goroutine logs after test completion.
// This option suppresses those panics and drops any subsequent logs.
func MuteAfterTestCompletion() LoggerOption {
	return loggerOptionFunc(func(opts *loggerOptions) {
		opts.muteAfterTestCompletion = true
	})
}

// NewLogger builds a new Logger that logs all messages to the given
// testing.TB.
//
//	logger := zaptest.NewLogger(t)
//
// Use this with a *testing.T or *testing.B to get logs which get printed only
// if a test fails or if you ran go test -v.
//
// The returned logger defaults to logging debug level messages and above.
// This may be changed by passing a zaptest.Level during construction.
//
//	logger := zaptest.NewLogger(t, zaptest.Level(zap.WarnLevel))
//
// You may also pass zap.Option's to customize test logger.
//
//	logger := zaptest.NewLogger(t, zaptest.WrapOptions(zap.AddCaller()))
func NewLogger(t TestingT, opts ...LoggerOption) *zap.Logger {
	cfg := loggerOptions{
		Level: zapcore.DebugLevel,
	}
	for _, o := range opts {
		o.applyLoggerOption(&cfg)
	}

	writer := NewTestingWriter(t)
	if cfg.muteAfterTestCompletion {
		writer = writer.WithMuteAfterTestCompletion(true)
	}
	zapOptions := []zap.Option{
		// Send zap errors to the same writer and mark the test as failed if
		// that happens.
		zap.ErrorOutput(writer.WithMarkFailed(true)),
	}
	zapOptions = append(zapOptions, cfg.zapOptions...)

	return zap.New(
		zapcore.NewCore(
			zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()),
			writer,
			cfg.Level,
		),
		zapOptions...,
	)
}

// TestingWriter is a WriteSyncer that writes to the given testing.TB.
type TestingWriter struct {
	t TestingT

	// If true, the test will be marked as failed if this TestingWriter is
	// ever used.
	markFailed bool

	// If true, log messages emitted after test completion will be ignored
	// instead of causing a panic.
	muteAfterTestCompletion bool

	// muted tracks whether a post-completion panic was encountered.
	muted *atomic.Bool
}

// NewTestingWriter builds a new TestingWriter that writes to the given
// testing.TB.
//
// Use this if you need more flexibility when creating *zap.Logger
// than zaptest.NewLogger() provides.
//
// E.g., if you want to use custom core with zaptest.TestingWriter:
//
//	encoder := newCustomEncoder()
//	writer := zaptest.NewTestingWriter(t)
//	level := zap.NewAtomicLevelAt(zapcore.DebugLevel)
//
//	core := newCustomCore(encoder, writer, level)
//
//	logger := zap.New(core, zap.AddCaller())
func NewTestingWriter(t TestingT) TestingWriter {
	return TestingWriter{
		t:     t,
		muted: new(atomic.Bool),
	}
}

// WithMarkFailed returns a copy of this TestingWriter with markFailed set to
// the provided value.
func (w TestingWriter) WithMarkFailed(v bool) TestingWriter {
	w.markFailed = v
	return w
}

// WithMuteAfterTestCompletion returns a copy of this TestingWriter with
// muteAfterTestCompletion set to the provided value.
func (w TestingWriter) WithMuteAfterTestCompletion(v bool) TestingWriter {
	w.muteAfterTestCompletion = v
	if w.muted == nil {
		w.muted = new(atomic.Bool)
	}
	return w
}

// Write writes bytes from p to the underlying testing.TB.
func (w TestingWriter) Write(p []byte) (n int, err error) {
	n = len(p)

	if w.muteAfterTestCompletion && w.muted != nil && w.muted.Load() {
		return n, nil
	}

	if w.muteAfterTestCompletion {
		defer func() {
			if r := recover(); r != nil {
				if isTestCompletedPanic(r) {
					if w.muted != nil {
						w.muted.Store(true)
					}
					return
				}
				panic(r)
			}
		}()
	}

	// Strip trailing newline because t.Log always adds one.
	p = bytes.TrimRight(p, "\n")

	// Note: t.Log is safe for concurrent use.
	w.t.Logf("%s", p)
	if w.markFailed {
		w.t.Fail()
	}

	return n, nil
}

func isTestCompletedPanic(r interface{}) bool {
	var s string
	switch v := r.(type) {
	case string:
		s = v
	case fmt.Stringer:
		s = v.String()
	default:
		return false
	}
	return strings.HasPrefix(s, "Log in goroutine after") ||
		strings.HasPrefix(s, "Fail in goroutine after")
}

// Sync commits the current contents (a no-op for TestingWriter).
func (w TestingWriter) Sync() error {
	return nil
}
