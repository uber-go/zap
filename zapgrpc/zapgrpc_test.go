// Copyright (c) 2016 Uber Technologies, Inc.
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

package zapgrpc

import (
	"fmt"
	"runtime"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	"github.com/stretchr/testify/require"
)

func TestLoggerInfoExpected(t *testing.T) {
	checkMessages(t, zapcore.DebugLevel, nil, zapcore.InfoLevel, []string{
		"hello",
		"s1s21 2 3s34s56",
		"hello world",
		"",
		"foo",
		"foo bar",
		"s1 s2 1 2 3 s3 4 s5 6",
		"hello",
		"s1s21 2 3s34s56",
		"hello world",
		"",
		"foo",
		"foo bar",
		"s1 s2 1 2 3 s3 4 s5 6",
	}, func(logger *Logger) {
		logger.Info("hello")
		logger.Info("s1", "s2", 1, 2, 3, "s3", 4, "s5", 6)
		logger.Infof("%s world", "hello")
		logger.Infoln()
		logger.Infoln("foo")
		logger.Infoln("foo", "bar")
		logger.Infoln("s1", "s2", 1, 2, 3, "s3", 4, "s5", 6)
		logger.Print("hello")
		logger.Print("s1", "s2", 1, 2, 3, "s3", 4, "s5", 6)
		logger.Printf("%s world", "hello")
		logger.Println()
		logger.Println("foo")
		logger.Println("foo", "bar")
		logger.Println("s1", "s2", 1, 2, 3, "s3", 4, "s5", 6)
	})
}

func TestLoggerDebugExpected(t *testing.T) {
	checkMessages(t, zapcore.DebugLevel, []Option{WithDebug()}, zapcore.DebugLevel, []string{
		"hello",
		"s1s21 2 3s34s56",
		"hello world",
		"",
		"foo",
		"foo bar",
		"s1 s2 1 2 3 s3 4 s5 6",
	}, func(logger *Logger) {
		logger.Print("hello")
		logger.Print("s1", "s2", 1, 2, 3, "s3", 4, "s5", 6)
		logger.Printf("%s world", "hello")
		logger.Println()
		logger.Println("foo")
		logger.Println("foo", "bar")
		logger.Println("s1", "s2", 1, 2, 3, "s3", 4, "s5", 6)
	})
}

func TestLoggerDebugSuppressed(t *testing.T) {
	checkMessages(t, zapcore.InfoLevel, []Option{WithDebug()}, zapcore.DebugLevel, nil, func(logger *Logger) {
		logger.Print("hello")
		logger.Printf("%s world", "hello")
		logger.Println()
		logger.Println("foo")
		logger.Println("foo", "bar")
	})
}

func TestLoggerWarningExpected(t *testing.T) {
	checkMessages(t, zapcore.DebugLevel, nil, zapcore.WarnLevel, []string{
		"hello",
		"s1s21 2 3s34s56",
		"hello world",
		"",
		"foo",
		"foo bar",
		"s1 s2 1 2 3 s3 4 s5 6",
	}, func(logger *Logger) {
		logger.Warning("hello")
		logger.Warning("s1", "s2", 1, 2, 3, "s3", 4, "s5", 6)
		logger.Warningf("%s world", "hello")
		logger.Warningln()
		logger.Warningln("foo")
		logger.Warningln("foo", "bar")
		logger.Warningln("s1", "s2", 1, 2, 3, "s3", 4, "s5", 6)
	})
}

func TestLoggerErrorExpected(t *testing.T) {
	checkMessages(t, zapcore.DebugLevel, nil, zapcore.ErrorLevel, []string{
		"hello",
		"s1s21 2 3s34s56",
		"hello world",
		"",
		"foo",
		"foo bar",
		"s1 s2 1 2 3 s3 4 s5 6",
	}, func(logger *Logger) {
		logger.Error("hello")
		logger.Error("s1", "s2", 1, 2, 3, "s3", 4, "s5", 6)
		logger.Errorf("%s world", "hello")
		logger.Errorln()
		logger.Errorln("foo")
		logger.Errorln("foo", "bar")
		logger.Errorln("s1", "s2", 1, 2, 3, "s3", 4, "s5", 6)
	})
}

func TestLoggerFatalExpected(t *testing.T) {
	checkMessages(t, zapcore.DebugLevel, nil, zapcore.FatalLevel, []string{
		"hello",
		"s1s21 2 3s34s56",
		"hello world",
		"",
		"foo",
		"foo bar",
		"s1 s2 1 2 3 s3 4 s5 6",
	}, func(logger *Logger) {
		logger.Fatal("hello")
		logger.Fatal("s1", "s2", 1, 2, 3, "s3", 4, "s5", 6)
		logger.Fatalf("%s world", "hello")
		logger.Fatalln()
		logger.Fatalln("foo")
		logger.Fatalln("foo", "bar")
		logger.Fatalln("s1", "s2", 1, 2, 3, "s3", 4, "s5", 6)
	})
}

func TestLoggerV(t *testing.T) {
	tests := []struct {
		zapLevel     zapcore.Level
		grpcEnabled  []int
		grpcDisabled []int
	}{
		{
			zapLevel:     zapcore.DebugLevel,
			grpcEnabled:  []int{grpcLvlInfo, grpcLvlWarn, grpcLvlError, grpcLvlFatal},
			grpcDisabled: []int{}, // everything is enabled, nothing is disabled
		},
		{
			zapLevel:     zapcore.InfoLevel,
			grpcEnabled:  []int{grpcLvlInfo, grpcLvlWarn, grpcLvlError, grpcLvlFatal},
			grpcDisabled: []int{}, // everything is enabled, nothing is disabled
		},
		{
			zapLevel:     zapcore.WarnLevel,
			grpcEnabled:  []int{grpcLvlWarn, grpcLvlError, grpcLvlFatal},
			grpcDisabled: []int{grpcLvlInfo},
		},
		{
			zapLevel:     zapcore.ErrorLevel,
			grpcEnabled:  []int{grpcLvlError, grpcLvlFatal},
			grpcDisabled: []int{grpcLvlInfo, grpcLvlWarn},
		},
		{
			zapLevel:     zapcore.DPanicLevel,
			grpcEnabled:  []int{grpcLvlFatal},
			grpcDisabled: []int{grpcLvlInfo, grpcLvlWarn, grpcLvlError},
		},
		{
			zapLevel:     zapcore.PanicLevel,
			grpcEnabled:  []int{grpcLvlFatal},
			grpcDisabled: []int{grpcLvlInfo, grpcLvlWarn, grpcLvlError},
		},
		{
			zapLevel:     zapcore.FatalLevel,
			grpcEnabled:  []int{grpcLvlFatal},
			grpcDisabled: []int{grpcLvlInfo, grpcLvlWarn, grpcLvlError},
		},
	}
	for _, tst := range tests {
		for _, grpcLvl := range tst.grpcEnabled {
			t.Run(fmt.Sprintf("enabled %s %d", tst.zapLevel, grpcLvl), func(t *testing.T) {
				checkLevel(t, tst.zapLevel, true, func(logger *Logger) bool {
					return logger.V(grpcLvl)
				})
			})
		}
		for _, grpcLvl := range tst.grpcDisabled {
			t.Run(fmt.Sprintf("disabled %s %d", tst.zapLevel, grpcLvl), func(t *testing.T) {
				checkLevel(t, tst.zapLevel, false, func(logger *Logger) bool {
					return logger.V(grpcLvl)
				})
			})
		}
	}
}

func checkLevel(
	t testing.TB,
	enab zapcore.LevelEnabler,
	expectedBool bool,
	f func(*Logger) bool,
) {
	withLogger(enab, nil, func(logger *Logger, observedLogs *observer.ObservedLogs) {
		actualBool := f(logger)
		if expectedBool {
			require.True(t, actualBool)
		} else {
			require.False(t, actualBool)
		}
	})
}

func checkMessages(
	t testing.TB,
	enab zapcore.LevelEnabler,
	opts []Option,
	expectedLevel zapcore.Level,
	expectedMessages []string,
	f func(*Logger),
) {
	if expectedLevel == zapcore.FatalLevel {
		expectedLevel = zapcore.WarnLevel
	}
	withLogger(enab, opts, func(logger *Logger, observedLogs *observer.ObservedLogs) {
		f(logger)
		logEntries := observedLogs.All()
		require.Equal(t, len(expectedMessages), len(logEntries))
		for i, logEntry := range logEntries {
			require.Equal(t, expectedLevel, logEntry.Level)
			require.Equal(t, expectedMessages[i], logEntry.Message)
		}
	})
}

func withLogger(
	enab zapcore.LevelEnabler,
	opts []Option,
	f func(*Logger, *observer.ObservedLogs),
) {
	core, observedLogs := observer.New(enab)
	f(NewLogger(zap.New(core), append(opts, withWarn())...), observedLogs)
}

func TestLoggerDepthExpected(t *testing.T) {
	checkMessages(t, zapcore.DebugLevel, nil, zapcore.InfoLevel, []string{
		"hello",
		"s1 s2 1 2 3 s3 4 s5 6",
		"",
	}, func(logger *Logger) {
		logger.InfoDepth(0, "hello")
		logger.InfoDepth(0, "s1", "s2", 1, 2, 3, "s3", 4, "s5", 6)
		logger.InfoDepth(0)
	})

	checkMessages(t, zapcore.DebugLevel, nil, zapcore.WarnLevel, []string{
		"hello",
		"s1 s2 1 2 3 s3 4 s5 6",
		"",
	}, func(logger *Logger) {
		logger.WarningDepth(0, "hello")
		logger.WarningDepth(0, "s1", "s2", 1, 2, 3, "s3", 4, "s5", 6)
		logger.WarningDepth(0)
	})

	checkMessages(t, zapcore.DebugLevel, nil, zapcore.ErrorLevel, []string{
		"hello",
		"s1 s2 1 2 3 s3 4 s5 6",
		"",
	}, func(logger *Logger) {
		logger.ErrorDepth(0, "hello")
		logger.ErrorDepth(0, "s1", "s2", 1, 2, 3, "s3", 4, "s5", 6)
		logger.ErrorDepth(0)
	})

	checkMessages(t, zapcore.DebugLevel, nil, zapcore.FatalLevel, []string{
		"hello",
		"s1 s2 1 2 3 s3 4 s5 6",
		"",
	}, func(logger *Logger) {
		logger.FatalDepth(0, "hello")
		logger.FatalDepth(0, "s1", "s2", 1, 2, 3, "s3", 4, "s5", 6)
		logger.FatalDepth(0)
	})
}

func TestLoggerDepthSuppressed(t *testing.T) {
	checkMessages(t, zapcore.WarnLevel, nil, zapcore.InfoLevel, nil, func(logger *Logger) {
		logger.InfoDepth(0, "hello")
	})
	checkMessages(t, zapcore.ErrorLevel, nil, zapcore.WarnLevel, nil, func(logger *Logger) {
		logger.WarningDepth(0, "hello")
	})
	checkMessages(t, zapcore.FatalLevel, nil, zapcore.ErrorLevel, nil, func(logger *Logger) {
		logger.ErrorDepth(0, "hello")
	})
}

func TestLoggerDepthCallerAttribution(t *testing.T) {
	core, observedLogs := observer.New(zapcore.DebugLevel)
	logger := NewLogger(zap.New(core, zap.AddCaller()), withWarn())

	logAtDepth := func(depth int, logFunc func(int, ...interface{}), msg string) {
		logFunc(depth, msg)
	}

	wrapper1 := func(depth int, logFunc func(int, ...interface{}), msg string) {
		logAtDepth(depth, logFunc, msg)
	}

	wrapper2 := func(depth int, logFunc func(int, ...interface{}), msg string) {
		wrapper1(depth, logFunc, msg)
	}

	tests := []struct {
		name    string
		logFunc func(int, ...interface{})
		lvl     zapcore.Level
	}{
		{"InfoDepth", logger.InfoDepth, zapcore.InfoLevel},
		{"WarningDepth", logger.WarningDepth, zapcore.WarnLevel},
		{"ErrorDepth", logger.ErrorDepth, zapcore.ErrorLevel},
		{"FatalDepth", logger.FatalDepth, zapcore.WarnLevel},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			observedLogs.TakeAll()

			// Depth 0: direct caller
			_, _, line0, _ := runtime.Caller(0)
			tt.logFunc(0, "depth 0") // expected at line0 + 1

			// Depth 1: skips logAtDepth
			_, _, line1, _ := runtime.Caller(0)
			logAtDepth(1, tt.logFunc, "depth 1") // expected at line1 + 1

			// Depth 2: skips logAtDepth and wrapper1
			_, _, line2, _ := runtime.Caller(0)
			wrapper1(2, tt.logFunc, "depth 2") // expected at line2 + 1

			// Depth 3: skips logAtDepth, wrapper1, and wrapper2
			_, _, line3, _ := runtime.Caller(0)
			wrapper2(3, tt.logFunc, "depth 3") // expected at line3 + 1

			logs := observedLogs.TakeAll()
			require.Len(t, logs, 4)

			for _, entry := range logs {
				require.Equal(t, tt.lvl, entry.Level)
				require.True(t, entry.Caller.Defined, "Caller must be defined")
				require.Contains(t, entry.Caller.File, "zapgrpc_test.go")
			}

			require.Equal(t, line0+1, logs[0].Caller.Line)
			require.Equal(t, line1+1, logs[1].Caller.Line)
			require.Equal(t, line2+1, logs[2].Caller.Line)
			require.Equal(t, line3+1, logs[3].Caller.Line)
		})
	}
}
