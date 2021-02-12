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
