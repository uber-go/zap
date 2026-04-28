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
	// Per grpclog.LoggerV2.V, V(l) reports whether verbosity level l is at
	// least the logger's configured verbose level. The Logger stores that
	// level via WithVerbosity (default 0), independent of zap's severity.
	tests := []struct {
		verbosity int
		enabled   []int
		disabled  []int
	}{
		{verbosity: 0, enabled: []int{0}, disabled: []int{1, 2, 3}},
		{verbosity: 1, enabled: []int{0, 1}, disabled: []int{2, 3}},
		{verbosity: 3, enabled: []int{0, 1, 2, 3}, disabled: []int{4}},
	}
	for _, tst := range tests {
		for _, l := range tst.enabled {
			t.Run(fmt.Sprintf("enabled verbosity=%d l=%d", tst.verbosity, l), func(t *testing.T) {
				logger := NewLogger(zap.NewNop(), WithVerbosity(tst.verbosity))
				if !logger.V(l) {
					t.Fatalf("V(%d) = false, want true at verbosity %d", l, tst.verbosity)
				}
			})
		}
		for _, l := range tst.disabled {
			t.Run(fmt.Sprintf("disabled verbosity=%d l=%d", tst.verbosity, l), func(t *testing.T) {
				logger := NewLogger(zap.NewNop(), WithVerbosity(tst.verbosity))
				if logger.V(l) {
					t.Fatalf("V(%d) = true, want false at verbosity %d", l, tst.verbosity)
				}
			})
		}
	}
}

func TestLoggerVDefaultVerbosity(t *testing.T) {
	// Without WithVerbosity, only V(0) is enabled, independent of zap level.
	for _, lvl := range []zapcore.Level{zapcore.DebugLevel, zapcore.ErrorLevel} {
		core, _ := observer.New(lvl)
		logger := NewLogger(zap.New(core))
		if !logger.V(0) {
			t.Fatalf("V(0) = false at zap level %s, want true", lvl)
		}
		if logger.V(1) {
			t.Fatalf("V(1) = true at zap level %s, want false (default verbosity 0)", lvl)
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
