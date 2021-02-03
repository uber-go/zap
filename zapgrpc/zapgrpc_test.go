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
		"hello world",
		"",
		"foo",
		"foo bar",
		"hello",
		"hello world",
		"",
		"foo",
		"foo bar",
	}, func(logger *Logger) {
		logger.Info("hello")
		logger.Infof("%s world", "hello")
		logger.Infoln()
		logger.Infoln("foo")
		logger.Infoln("foo", "bar")
		logger.Print("hello")
		logger.Printf("%s world", "hello")
		logger.Println()
		logger.Println("foo")
		logger.Println("foo", "bar")
	})
}

func TestLoggerDebugExpected(t *testing.T) {
	checkMessages(t, zapcore.DebugLevel, []Option{WithDebug()}, zapcore.DebugLevel, []string{
		"hello",
		"hello world",
		"",
		"foo",
		"foo bar",
	}, func(logger *Logger) {
		logger.Print("hello")
		logger.Printf("%s world", "hello")
		logger.Println()
		logger.Println("foo")
		logger.Println("foo", "bar")
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
		"hello world",
		"",
		"foo",
		"foo bar",
	}, func(logger *Logger) {
		logger.Warning("hello")
		logger.Warningf("%s world", "hello")
		logger.Warningln()
		logger.Warningln("foo")
		logger.Warningln("foo", "bar")
	})
}

func TestLoggerErrorExpected(t *testing.T) {
	checkMessages(t, zapcore.DebugLevel, nil, zapcore.ErrorLevel, []string{
		"hello",
		"hello world",
		"",
		"foo",
		"foo bar",
	}, func(logger *Logger) {
		logger.Error("hello")
		logger.Errorf("%s world", "hello")
		logger.Errorln()
		logger.Errorln("foo")
		logger.Errorln("foo", "bar")
	})
}

func TestLoggerFatalExpected(t *testing.T) {
	checkMessages(t, zapcore.DebugLevel, nil, zapcore.FatalLevel, []string{
		"hello",
		"hello world",
		"",
		"foo",
		"foo bar",
	}, func(logger *Logger) {
		logger.Fatal("hello")
		logger.Fatalf("%s world", "hello")
		logger.Fatalln()
		logger.Fatalln("foo")
		logger.Fatalln("foo", "bar")
	})
}

func TestLoggerVTrueExpected(t *testing.T) {
	enabled := map[zapcore.Level][]int{
		zapcore.DebugLevel: {
			grpcLvlInfo, grpcLvlWarn, grpcLvlError, grpcLvlFatal,
		},
		zapcore.InfoLevel: {
			grpcLvlInfo, grpcLvlWarn, grpcLvlError, grpcLvlFatal,
		},
		zapcore.WarnLevel: {
			grpcLvlWarn, grpcLvlError, grpcLvlFatal,
		},
		zapcore.ErrorLevel: {
			grpcLvlError, grpcLvlFatal,
		},
		zapcore.DPanicLevel: {
			grpcLvlFatal,
		},
		zapcore.PanicLevel: {
			grpcLvlFatal,
		},
		zapcore.FatalLevel: {
			grpcLvlFatal,
		},
	}
	for zapLvl, grpcLvls := range enabled {
		for _, grpcLvl := range grpcLvls {
			t.Run(fmt.Sprintf("%s %d", zapLvl, grpcLvl), func(t *testing.T) {
				checkLevel(t, zapLvl, true, func(logger *Logger) bool {
					return logger.V(grpcLvl)
				})
			})
		}
	}
}

func TestLoggerVFalseExpected(t *testing.T) {
	disabled := map[zapcore.Level][]int{
		zapcore.DebugLevel: {
			// everything is enabled, nothing is disabled
		},
		zapcore.InfoLevel: {
			// everything is enabled, nothing is disabled
		},
		zapcore.WarnLevel: {
			grpcLvlInfo,
		},
		zapcore.ErrorLevel: {
			grpcLvlInfo, grpcLvlWarn,
		},
		zapcore.DPanicLevel: {
			grpcLvlInfo, grpcLvlWarn, grpcLvlError,
		},
		zapcore.PanicLevel: {
			grpcLvlInfo, grpcLvlWarn, grpcLvlError,
		},
		zapcore.FatalLevel: {
			grpcLvlInfo, grpcLvlWarn, grpcLvlError,
		},
	}
	for zapLvl, grpcLvls := range disabled {
		for _, grpcLvl := range grpcLvls {
			t.Run(fmt.Sprintf("%s %d", zapLvl, grpcLvl), func(t *testing.T) {
				checkLevel(t, zapLvl, false, func(logger *Logger) bool {
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

// withWarn redirects the fatal level to the warn level, which makes testing
// easier.
func withWarn() Option {
	return optionFunc(func(logger *Logger) {
		logger.fatalToWarn = true
	})
}
