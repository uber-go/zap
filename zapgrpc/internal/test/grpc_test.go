// Copyright (c) 2021 Uber Technologies, Inc.
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

package grpc

import (
	"fmt"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zapgrpc"
	"go.uber.org/zap/zaptest"
	"go.uber.org/zap/zaptest/observer"
	"google.golang.org/grpc/grpclog"
)

func TestLogger(t *testing.T) {
	core, observedLogs := observer.New(zapcore.InfoLevel)
	zlog := zap.New(core)

	grpclog.SetLogger(zapgrpc.NewLogger(zlog))

	grpclog.Print("hello from grpc")

	logs := observedLogs.TakeAll()
	require.Len(t, logs, 1, "Expected one log entry.")
	entry := logs[0]

	assert.Equal(t, zapcore.InfoLevel, entry.Level,
		"Log entry level did not match.")
	assert.Equal(t, "hello from grpc", entry.Message,
		"Log entry message did not match.")
}

func TestLoggerV2(t *testing.T) {
	core, observedLogs := observer.New(zapcore.InfoLevel)
	zlog := zap.New(core)

	grpclog.SetLoggerV2(zapgrpc.NewLogger(zlog))

	grpclog.Info("hello from grpc")

	logs := observedLogs.TakeAll()
	require.Len(t, logs, 1, "Expected one log entry.")
	entry := logs[0]

	assert.Equal(t, zapcore.InfoLevel, entry.Level,
		"Log entry level did not match.")
	assert.Equal(t, "hello from grpc", entry.Message,
		"Log entry message did not match.")
}

func TestDepthLogger(t *testing.T) {
	defer grpclog.SetLoggerV2(zapgrpc.NewLogger(zap.NewNop()))

	comp := grpclog.Component("test")

	args := []interface{}{"message", "param"}
	cases := []struct {
		name  string
		fn    func(...interface{})
		level zapcore.Level
	}{
		{name: "Info", fn: grpclog.Info, level: zap.InfoLevel},
		{name: "Infoln", fn: grpclog.Infoln, level: zap.InfoLevel},
		{name: "comp.Info", fn: comp.Info, level: zap.InfoLevel},
		{name: "Warning", fn: grpclog.Warning, level: zap.WarnLevel},
		{name: "Warningln", fn: grpclog.Warningln, level: zap.WarnLevel},
		{name: "comp.Warning", fn: comp.Warning, level: zap.WarnLevel},
		{name: "Error", fn: grpclog.Error, level: zap.ErrorLevel},
		{name: "Errorln", fn: grpclog.Errorln, level: zap.ErrorLevel},
		{name: "comp.Error", fn: comp.Error, level: zap.ErrorLevel},
		{name: "Fatal", fn: grpclog.Fatal, level: zap.FatalLevel},
		{name: "Fatalln", fn: grpclog.Fatalln, level: zap.FatalLevel},
		{name: "comp.Fatal", fn: comp.Fatal, level: zap.FatalLevel},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			called := false
			logger := zaptest.NewLogger(t, zaptest.WrapOptions(
				zap.AddCaller(),
				zap.WithFatalHook(zapcore.WriteThenPanic),
				zap.Hooks(func(entry zapcore.Entry) error {
					called = true
					require.Equal(t, c.level, entry.Level)
					prefix := ""
					if strings.HasPrefix(c.name, "comp") {
						prefix = "[test]"
					}
					if strings.HasSuffix(c.name, "ln") {
						require.Equal(t, prefix+sprintln(args), entry.Message)
					} else {
						require.Equal(t, prefix+fmt.Sprint(args...), entry.Message)
					}
					_, file, _, _ := runtime.Caller(0)
					require.Equal(t, file, entry.Caller.File, entry.Caller)
					return nil
				}),
			))
			grpclog.SetLoggerV2(zapgrpc.NewLogger(logger))

			if c.level != zap.FatalLevel {
				c.fn(args...)
			} else {
				require.Panics(t, func() {
					c.fn(args...)
				})
			}
			require.True(t, called, "hook not called")
		})
	}
}

func Test_zapGrpcLogger_V(t *testing.T) {
	const (
		// The default verbosity level.
		// See https://github.com/grpc/grpc-go/blob/8ab16ef276a33df4cdb106446eeff40ff56a6928/grpclog/loggerv2.go#L108.
		normal = 0

		// Currently the only level of "being verbose".
		// For example https://github.com/grpc/grpc-go/blob/8ab16ef276a33df4cdb106446eeff40ff56a6928/grpclog/grpclog.go#L21.
		verbose = 2

		// As is mentioned in https://github.com/grpc/grpc-go/blob/8ab16ef276a33df4cdb106446eeff40ff56a6928/README.md#how-to-turn-on-logging,
		// though currently not being used in the code.
		extremelyVerbose = 99
	)

	grpclog.SetLoggerV2(zapgrpc.NewLogger(zap.NewNop(), zapgrpc.WithVerbosity(2)))
	assert.True(t, grpclog.V(normal))
	assert.True(t, grpclog.V(verbose))
	assert.False(t, grpclog.V(extremelyVerbose))
}

func sprintln(args []interface{}) string {
	s := fmt.Sprintln(args...)
	// Drop the new line character added by Sprintln
	return s[:len(s)-1]
}
