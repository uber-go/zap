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
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
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
		require.Panics(t, func() { logger.Fatal("hello") })
		require.Panics(t, func() { logger.Fatal("s1", "s2", 1, 2, 3, "s3", 4, "s5", 6) })
		require.Panics(t, func() { logger.Fatalf("%s world", "hello") })
		require.Panics(t, func() { logger.Fatalln() })
		require.Panics(t, func() { logger.Fatalln("foo") })
		require.Panics(t, func() { logger.Fatalln("foo", "bar") })
		require.Panics(t, func() { logger.Fatalln("s1", "s2", 1, 2, 3, "s3", 4, "s5", 6) })
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
	f(NewLogger(zap.New(core, zap.WithFatalHook(zapcore.WriteThenPanic)), opts...), observedLogs)
}
