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
	}, func(logger *Logger) {
		logger.Print("hello")
		logger.Printf("world")
		logger.Println("foo")
	})
}

func TestLoggerDebugExpected(t *testing.T) {
	checkMessages(t, zapcore.DebugLevel, []LoggerOption{WithDebug()}, zapcore.DebugLevel, []string{
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
	checkMessages(t, zapcore.InfoLevel, []LoggerOption{WithDebug()}, zapcore.DebugLevel, nil, func(logger *Logger) {
		logger.Print("hello")
		logger.Printf("world")
		logger.Println("foo")
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

func checkMessages(
	t testing.TB,
	levelEnabler zapcore.LevelEnabler,
	loggerOptions []LoggerOption,
	expectedLevel zapcore.Level,
	expectedMessages []string,
	f func(*Logger),
) {
	if expectedLevel == zapcore.FatalLevel {
		expectedLevel = zapcore.WarnLevel
	}
	withLogger(levelEnabler, loggerOptions, func(logger *Logger, observedLogs *observer.ObservedLogs) {
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
	levelEnabler zapcore.LevelEnabler,
	loggerOptions []LoggerOption,
	f func(*Logger, *observer.ObservedLogs),
) {
	core, observedLogs := observer.New(levelEnabler)
	f(NewLogger(zap.New(core), append(loggerOptions, withWarn())...), observedLogs)
}

// withWarn redirects the fatal level to the warn level.
//
// This is used for testing.
func withWarn() LoggerOption {
	return func(logger *Logger) {
		logger.fatalFunc = (*zap.SugaredLogger).Warn
		logger.fatalfFunc = (*zap.SugaredLogger).Warnf
	}
}
