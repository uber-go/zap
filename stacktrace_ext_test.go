// Copyright (c) 2016, 2017 Uber Technologies, Inc.
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

package zap_test

import (
	"bytes"
	"strings"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStacktraceBeginsAtLogCallSite(t *testing.T) {
	withLogger(t, func(logger *zap.Logger, out *bytes.Buffer) {
		logger.Error("test log")
		logger.Sugar().Error("sugar test log")

		frames := strings.Split(out.String(), "\n")

		// stack from logger.Error() call
		stack := []string{
			"go.uber.org/zap_test.TestStacktraceBeginsAtLogCallSite.func1",
			"go.uber.org/zap_test.withLogger",
			"go.uber.org/zap_test.TestStacktraceBeginsAtLogCallSite",
			"testing.tRunner",
		}

		require.True(t, len(frames) > len(stack)*2+1)

		for idx, fn := range stack {
			assert.Equalf(t, fn, frames[1+2*idx], "frame %v unexpected", idx)
		}

		// stack from logger.Sugar().Error() call
		frames = frames[2*len(stack)+1:]

		stack = []string{
			"go.uber.org/zap_test.TestStacktraceBeginsAtLogCallSite.func1",
			"go.uber.org/zap_test.withLogger",
			"go.uber.org/zap_test.TestStacktraceBeginsAtLogCallSite",
			"testing.tRunner",
		}

		require.True(t, len(frames) > len(stack)*2)

		for idx, fn := range stack {
			assert.Equalf(t, fn, frames[1+2*idx], "sugared frame %v unexpected", idx)
		}
	})
}

func TestStacktraceHonorsCallerSkip(t *testing.T) {
	withLogger(t, func(logger *zap.Logger, out *bytes.Buffer) {
		func() {
			logger.Error("test log")
		}()

		frames := strings.Split(out.String(), "\n")

		// stack from logger.Error() call not including enclosing func
		stack := []string{
			"go.uber.org/zap_test.TestStacktraceHonorsCallerSkip.func1",
			"go.uber.org/zap_test.withLogger",
			"go.uber.org/zap_test.TestStacktraceHonorsCallerSkip",
			"testing.tRunner",
		}

		require.True(t, len(frames) > len(stack)*2)

		for idx, fn := range stack {
			assert.Equalf(t, fn, frames[1+2*idx], "frame %v unexpected", idx)
		}
	}, zap.AddCallerSkip(1))
}

func TestStacktraceIncludesZapFramesAfterCallSite(t *testing.T) {
	withLogger(t, func(logger *zap.Logger, out *bytes.Buffer) {
		marshal := func(enc zapcore.ObjectEncoder) error {
			logger.Warn("marshal caused warn")
			enc.AddString("f", "v")
			return nil
		}
		logger.Error("test log", zap.Object("obj", zapcore.ObjectMarshalerFunc(marshal)))

		logs := out.String()
		frames := strings.Split(logs, "\n")

		stack := []string{
			"go.uber.org/zap_test.TestStacktraceIncludesZapFramesAfterCallSite.func1.1",
			"go.uber.org/zap/zapcore.ObjectMarshalerFunc.MarshalLogObject",
			"go.uber.org/zap/zapcore.(*jsonEncoder).AppendObject",
			"go.uber.org/zap/zapcore.(*jsonEncoder).AddObject",
			"go.uber.org/zap/zapcore.Field.AddTo",
			"go.uber.org/zap/zapcore.addFields",
			"go.uber.org/zap/zapcore.consoleEncoder.writeContext",
			"go.uber.org/zap/zapcore.consoleEncoder.EncodeEntry",
			"go.uber.org/zap/zapcore.(*ioCore).Write",
			"go.uber.org/zap/zapcore.(*CheckedEntry).Write",
			"go.uber.org/zap.(*Logger).Error",
			"go.uber.org/zap_test.TestStacktraceIncludesZapFramesAfterCallSite.func1",
			"go.uber.org/zap_test.withLogger",
			"go.uber.org/zap_test.TestStacktraceIncludesZapFramesAfterCallSite",
			"testing.tRunner",
		}

		require.True(t, len(frames) > len(stack)*2)

		for idx, fn := range stack {
			assert.Equalf(t, fn, frames[1+2*idx], "frame %v unexpected", idx)
		}
	})
}

// withLogger sets up a logger with a real encoder set up, so that any marshal functions are called.
// The inbuilt observer does not call Marshal for objects/arrays, which we need for some tests.
func withLogger(t *testing.T, fn func(logger *zap.Logger, out *bytes.Buffer), opts ...zap.Option) {
	buf := &bytes.Buffer{}
	encoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
	core := zapcore.NewCore(encoder, zapcore.AddSync(buf), zapcore.DebugLevel)

	opts = append([]zap.Option{zap.AddStacktrace(zap.DebugLevel)}, opts...)
	logger := zap.New(core, opts...)

	fn(logger, buf)
}
