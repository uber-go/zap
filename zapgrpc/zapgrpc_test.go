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

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	"github.com/stretchr/testify/require"
)

func TestLoggerInfoExpected(t *testing.T) {
	checkMessages(t, zapcore.DebugLevel, nil, zapcore.InfoLevel, []string{
		"hello",
		"world",
		"foo",
		"hello",
		"world",
		"foo",
	}, func(logger *Logger) {
		logger.Info("hello")
		logger.Infof("world")
		logger.Infoln("foo")
		logger.Print("hello")
		logger.Printf("world")
		logger.Println("foo")
	})
}

func TestLoggerDebugExpected(t *testing.T) {
	checkMessages(t, zapcore.DebugLevel, []Option{WithDebug()}, zapcore.DebugLevel, []string{
		"hello",
		"world",
		"foo",
	}, func(logger *Logger) {
		logger.Print("hello")
		logger.Printf("world")
		logger.Println("foo")
	})
}

func TestLoggerDebugSuppressed(t *testing.T) {
	checkMessages(t, zapcore.InfoLevel, []Option{WithDebug()}, zapcore.DebugLevel, nil, func(logger *Logger) {
		logger.Print("hello")
		logger.Printf("world")
		logger.Println("foo")
	})
}

func TestLoggerWarnExpected(t *testing.T) {
	checkMessages(t, zapcore.DebugLevel, nil, zapcore.WarnLevel, []string{
		"hello",
		"world",
		"foo",
	}, func(logger *Logger) {
		logger.Warning("hello")
		logger.Warningf("world")
		logger.Warningln("foo")
	})
}

func TestLoggerErrorExpected(t *testing.T) {
	checkMessages(t, zapcore.DebugLevel, nil, zapcore.ErrorLevel, []string{
		"hello",
		"world",
		"foo",
	}, func(logger *Logger) {
		logger.Error("hello")
		logger.Errorf("world")
		logger.Errorln("foo")
	})
}

func TestLoggerFatalExpected(t *testing.T) {
	checkMessages(t, zapcore.DebugLevel, nil, zapcore.FatalLevel, []string{
		"hello",
		"world",
		"foo",
	}, func(logger *Logger) {
		logger.Fatal("hello")
		logger.Fatalf("world")
		logger.Fatalln("foo")
	})
}

func TestLoggerTrueExpected(t *testing.T) {
	checkLevel(t, zapcore.FatalLevel, false, func(logger *Logger) bool {
		return logger.V(6)
	})
}

func TestLoggerFalseExpected(t *testing.T) {
	checkLevel(t, zapcore.FatalLevel, true, func(logger *Logger) bool {
		return logger.V(0)
	})
}

func checkLevel(
	t testing.TB,
	enab zapcore.LevelEnabler,
	expectedBool bool,
	f func(*Logger) bool,
) {
	withLogger(enab, nil, func(logger *Logger, observedLogs *observer.ObservedLogs) {
		actualBool := f(logger)
		require.Equal(t, expectedBool, actualBool)
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

// withWarn redirects the fatal level to the warn level, which makes testing
// easier.
func withWarn() Option {
	return optionFunc(func(logger *Logger) {
		logger.fatal = (*zap.SugaredLogger).Warn
		logger.fatalf = (*zap.SugaredLogger).Warnf
	})
}
