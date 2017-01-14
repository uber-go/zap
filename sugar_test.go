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

package zap

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"go.uber.org/zap/internal/exit"
	"go.uber.org/zap/internal/observer"
	"go.uber.org/zap/zapcore"
)

func TestSugarWith(t *testing.T) {
	ignored := observer.LoggedEntry{
		Entry:   zapcore.Entry{Level: DPanicLevel, Message: _oddNumberErrMsg},
		Context: []zapcore.Field{Any("ignored", "should ignore")},
	}

	tests := []struct {
		args     []interface{}
		expected []zapcore.Field
	}{
		{nil, []zapcore.Field{}},
		{[]interface{}{}, []zapcore.Field{}},
		{[]interface{}{"foo", 42, true, "bar"}, []zapcore.Field{Int("foo", 42), String("true", "bar")}},
	}

	for _, tt := range tests {
		withSugar(t, DebugLevel, nil, func(logger *SugaredLogger, logs *observer.ObservedLogs) {
			logger.With(tt.args...).Info("")
			output := logs.AllUntimed()
			assert.Equal(t, 1, len(output), "Expected only one message to be logged.")
			assert.Equal(t, tt.expected, output[0].Context, "Unexpected message context.")
		})

		withSugar(t, DebugLevel, nil, func(logger *SugaredLogger, logs *observer.ObservedLogs) {
			oddArgs := append(tt.args, "should ignore")
			logger.With(oddArgs...).Info("")
			output := logs.AllUntimed()
			assert.Equal(t, 2, len(output), "Expected an error to be logged along with the intended message.")
			assert.Equal(t, ignored, output[0], "Expected the first message to be an error.")
			assert.Equal(t, tt.expected, output[1].Context, "Unexpected context on intended message.")
		})
	}
}

func TestSugarWithFields(t *testing.T) {
	tests := [][]zapcore.Field{
		{},
		{String("foo", "bar"), Int("baz", 42)},
	}
	for _, fields := range tests {
		withSugar(t, DebugLevel, nil, func(logger *SugaredLogger, logs *observer.ObservedLogs) {
			logger.WithFields(fields...).Info("")
			output := logs.AllUntimed()
			assert.Equal(t, 1, len(output), "Expected only one message to be logged.")
			assert.Equal(t, fields, output[0].Context, "Unexpected message context.")
		})
	}
}

type stringerF func() string

func (f stringerF) String() string { return f() }

func TestSugarStructuredLogging(t *testing.T) {
	tests := []struct {
		msg       interface{}
		expectMsg string
	}{
		{"foo", "foo"},
		{true, "true"},
		{stringerF(func() string { return "hello" }), "hello"},
	}

	// Common to all test cases.
	context := []interface{}{"foo", "bar"}
	extra := []interface{}{true, false}
	expectedFields := []zapcore.Field{String("foo", "bar"), Bool("true", false)}

	for _, tt := range tests {
		withSugar(t, DebugLevel, nil, func(logger *SugaredLogger, logs *observer.ObservedLogs) {
			logger.With(context...).Debug(tt.msg, extra...)
			logger.With(context...).Info(tt.msg, extra...)
			logger.With(context...).Warn(tt.msg, extra...)
			logger.With(context...).Error(tt.msg, extra...)
			logger.With(context...).DPanic(tt.msg, extra...)

			expected := make([]observer.LoggedEntry, 5)
			for i, lvl := range []zapcore.Level{DebugLevel, InfoLevel, WarnLevel, ErrorLevel, DPanicLevel} {
				expected[i] = observer.LoggedEntry{
					Entry:   zapcore.Entry{Message: tt.expectMsg, Level: lvl},
					Context: expectedFields,
				}
			}
			assert.Equal(t, expected, logs.AllUntimed(), "Unexpected log output.")
		})
	}
}

func TestSugarFormattedLogging(t *testing.T) {
	tests := []struct {
		format string
		args   []interface{}
		expect string
	}{
		{"", nil, ""},
		{"foo", nil, "foo"},
		{"", []interface{}{"foo"}, "%!(EXTRA string=foo)"},
	}

	// Common to all test cases.
	context := []interface{}{"foo", "bar"}
	expectedFields := []zapcore.Field{String("foo", "bar")}

	for _, tt := range tests {
		withSugar(t, DebugLevel, nil, func(logger *SugaredLogger, logs *observer.ObservedLogs) {
			logger.With(context...).Debugf(tt.format, tt.args...)
			logger.With(context...).Infof(tt.format, tt.args...)
			logger.With(context...).Warnf(tt.format, tt.args...)
			logger.With(context...).Errorf(tt.format, tt.args...)
			logger.With(context...).DPanicf(tt.format, tt.args...)

			expected := make([]observer.LoggedEntry, 5)
			for i, lvl := range []zapcore.Level{DebugLevel, InfoLevel, WarnLevel, ErrorLevel, DPanicLevel} {
				expected[i] = observer.LoggedEntry{
					Entry:   zapcore.Entry{Message: tt.expect, Level: lvl},
					Context: expectedFields,
				}
			}
			assert.Equal(t, expected, logs.AllUntimed(), "Unexpected log output.")
		})
	}
}

func TestSugarPanicLogging(t *testing.T) {
	tests := []struct {
		loggerLevel zapcore.Level
		f           func(*SugaredLogger)
		expectedMsg string
	}{
		{FatalLevel, func(s *SugaredLogger) { s.Panic("foo") }, ""},
		{PanicLevel, func(s *SugaredLogger) { s.Panic("foo") }, "foo"},
		{DebugLevel, func(s *SugaredLogger) { s.Panic("foo") }, "foo"},
		{FatalLevel, func(s *SugaredLogger) { s.Panicf("%s", "foo") }, ""},
		{PanicLevel, func(s *SugaredLogger) { s.Panicf("%s", "foo") }, "foo"},
		{DebugLevel, func(s *SugaredLogger) { s.Panicf("%s", "foo") }, "foo"},
	}

	for _, tt := range tests {
		withSugar(t, tt.loggerLevel, nil, func(sugar *SugaredLogger, logs *observer.ObservedLogs) {
			assert.Panics(t, func() { tt.f(sugar) }, "Expected panic-level logger calls to panic.")
			if tt.expectedMsg != "" {
				assert.Equal(t, []observer.LoggedEntry{{
					Context: []zapcore.Field{},
					Entry:   zapcore.Entry{Message: tt.expectedMsg, Level: PanicLevel},
				}}, logs.AllUntimed(), "Unexpected log output.")
			} else {
				assert.Equal(t, 0, logs.Len(), "Didn't expect any log output.")
			}
		})
	}
}

func TestSugarFatalLogging(t *testing.T) {
	tests := []struct {
		loggerLevel zapcore.Level
		f           func(*SugaredLogger)
		expectedMsg string
	}{
		{FatalLevel + 1, func(s *SugaredLogger) { s.Fatal("foo") }, ""},
		{FatalLevel, func(s *SugaredLogger) { s.Fatal("foo") }, "foo"},
		{DebugLevel, func(s *SugaredLogger) { s.Fatal("foo") }, "foo"},
		{FatalLevel + 1, func(s *SugaredLogger) { s.Fatalf("%s", "foo") }, ""},
		{FatalLevel, func(s *SugaredLogger) { s.Fatalf("%s", "foo") }, "foo"},
		{DebugLevel, func(s *SugaredLogger) { s.Fatalf("%s", "foo") }, "foo"},
	}

	for _, tt := range tests {
		withSugar(t, tt.loggerLevel, nil, func(sugar *SugaredLogger, logs *observer.ObservedLogs) {
			stub := exit.WithStub(func() { tt.f(sugar) })
			assert.True(t, stub.Exited, "Expected all calls to fatal logger methods to exit process.")
			if tt.expectedMsg != "" {
				assert.Equal(t, []observer.LoggedEntry{{
					Context: []zapcore.Field{},
					Entry:   zapcore.Entry{Message: tt.expectedMsg, Level: FatalLevel},
				}}, logs.AllUntimed(), "Unexpected log output.")
			} else {
				assert.Equal(t, 0, logs.Len(), "Didn't expect any log output.")
			}
		})
	}
}
