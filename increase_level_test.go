// Copyright (c) 2020 Uber Technologies, Inc.
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

package zap

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func newLoggedEntry(level zapcore.Level, msg string) observer.LoggedEntry {
	return observer.LoggedEntry{
		Entry:   zapcore.Entry{Level: level, Message: msg},
		Context: []zapcore.Field{},
	}
}

func TestIncreaseLevelTryDecrease(t *testing.T) {
	errorOut := &bytes.Buffer{}
	opts := []Option{
		ErrorOutput(zapcore.AddSync(errorOut)),
	}
	withLogger(t, WarnLevel, opts, func(logger *Logger, logs *observer.ObservedLogs) {
		logger.Warn("original warn log")

		debugLogger := logger.WithOptions(IncreaseLevel(DebugLevel))
		debugLogger.Debug("ignored debug log")
		debugLogger.Warn("increase level warn log")
		debugLogger.Error("increase level error log")

		assert.Equal(t, []observer.LoggedEntry{
			newLoggedEntry(WarnLevel, "original warn log"),
			newLoggedEntry(WarnLevel, "increase level warn log"),
			newLoggedEntry(ErrorLevel, "increase level error log"),
		}, logs.AllUntimed(), "unexpected logs")
		assert.Equal(t,
			`failed to IncreaseLevel: invalid increase level, as level "info" is allowed by increased level, but not by existing core`,
			errorOut.String(),
			"unexpected error output",
		)
	})
}

func TestIncreaseLevel(t *testing.T) {
	errorOut := &bytes.Buffer{}
	opts := []Option{
		ErrorOutput(zapcore.AddSync(errorOut)),
	}
	withLogger(t, WarnLevel, opts, func(logger *Logger, logs *observer.ObservedLogs) {
		logger.Warn("original warn log")

		errorLogger := logger.WithOptions(IncreaseLevel(ErrorLevel))
		errorLogger.Debug("ignored debug log")
		errorLogger.Warn("ignored warn log")
		errorLogger.Error("increase level error log")

		assert.Equal(t, []observer.LoggedEntry{
			newLoggedEntry(WarnLevel, "original warn log"),
			newLoggedEntry(ErrorLevel, "increase level error log"),
		}, logs.AllUntimed(), "unexpected logs")

		assert.Empty(t, errorOut.String(), "expect no error output")
	})
}
