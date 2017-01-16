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
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"

	"go.uber.org/zap/internal/exit"
	"go.uber.org/zap/internal/observer"
	"go.uber.org/zap/zapcore"
)

type sortedFields []zapcore.Field

func (sf sortedFields) Len() int           { return len(sf) }
func (sf sortedFields) Less(i, j int) bool { return sf[i].Key < sf[j].Key }
func (sf sortedFields) Swap(i, j int)      { sf[i], sf[j] = sf[j], sf[i] }

func TestSugarWith(t *testing.T) {
	tests := []struct {
		ctx      Ctx
		expected []zapcore.Field
	}{
		{nil, []zapcore.Field{}},
		{Ctx{}, []zapcore.Field{}},
		{Ctx{"foo": 42, "bar": "baz"}, []zapcore.Field{Int("foo", 42), String("bar", "baz")}},
	}

	for _, tt := range tests {
		withSugar(t, DebugLevel, nil, func(logger *SugaredLogger, logs *observer.ObservedLogs) {
			logger.With(tt.ctx).Info("")
			output := logs.AllUntimed()
			assert.Equal(t, 1, len(output), "Expected only one message to be logged.")
			sort.Sort(sortedFields(tt.expected))
			sort.Sort(sortedFields(output[0].Context))
			assert.Equal(t, tt.expected, output[0].Context, "Unexpected message context.")
		})
	}
}

type stringerF func() string

func (f stringerF) String() string { return f() }

func TestSugarUntemplatedLogging(t *testing.T) {
	context := Ctx{"foo": "bar"}
	extra := Ctx{"baz": "quux"}
	expectedCtx := []zapcore.Field{String("foo", "bar")}
	expectedCtxExtra := []zapcore.Field{String("foo", "bar"), String("baz", "quux")}

	withSugar(t, DebugLevel, nil, func(logger *SugaredLogger, logs *observer.ObservedLogs) {
		logger.With(context).Debug("msg")
		logger.With(context).DebugWith("msg", extra)
		logger.With(context).Info("msg")
		logger.With(context).InfoWith("msg", extra)
		logger.With(context).Warn("msg")
		logger.With(context).WarnWith("msg", extra)
		logger.With(context).Error("msg")
		logger.With(context).ErrorWith("msg", extra)
		logger.With(context).DPanic("msg")
		logger.With(context).DPanicWith("msg", extra)

		expectedLogs := make([]observer.LoggedEntry, 10)
		for i, lvl := range []zapcore.Level{DebugLevel, InfoLevel, WarnLevel, ErrorLevel, DPanicLevel} {
			expectedLogs[i*2] = observer.LoggedEntry{
				Entry:   zapcore.Entry{Message: "msg", Level: lvl},
				Context: expectedCtx,
			}
			expectedLogs[i*2+1] = observer.LoggedEntry{
				Entry:   zapcore.Entry{Message: "msg", Level: lvl},
				Context: expectedCtxExtra,
			}
		}
		assert.Equal(t, expectedLogs, logs.AllUntimed(), "Unexpected log output.")
	})
}

func TestSugarTemplatedLogging(t *testing.T) {
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
	context := Ctx{"foo": "bar"}
	expectedFields := []zapcore.Field{String("foo", "bar")}

	for _, tt := range tests {
		withSugar(t, DebugLevel, nil, func(logger *SugaredLogger, logs *observer.ObservedLogs) {
			logger.With(context).Debugf(tt.format, tt.args...)
			logger.With(context).Infof(tt.format, tt.args...)
			logger.With(context).Warnf(tt.format, tt.args...)
			logger.With(context).Errorf(tt.format, tt.args...)
			logger.With(context).DPanicf(tt.format, tt.args...)

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
		{FatalLevel, func(s *SugaredLogger) { s.PanicWith("foo", nil) }, ""},
		{PanicLevel, func(s *SugaredLogger) { s.PanicWith("foo", nil) }, "foo"},
		{DebugLevel, func(s *SugaredLogger) { s.PanicWith("foo", nil) }, "foo"},
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
		{FatalLevel + 1, func(s *SugaredLogger) { s.FatalWith("foo", nil) }, ""},
		{FatalLevel, func(s *SugaredLogger) { s.FatalWith("foo", nil) }, "foo"},
		{DebugLevel, func(s *SugaredLogger) { s.FatalWith("foo", nil) }, "foo"},
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
