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
	"errors"
	"fmt"
	"strconv"
	"testing"

	"go.uber.org/zap/internal/exit"
	"go.uber.org/zap/internal/ztest"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSugarWith(t *testing.T) {
	// Convenience functions to create expected error logs.
	ignored := func(msg interface{}) observer.LoggedEntry {
		return observer.LoggedEntry{
			Entry:   zapcore.Entry{Level: ErrorLevel, Message: _oddNumberErrMsg},
			Context: []Field{Any("ignored", msg)},
		}
	}
	nonString := func(pairs ...invalidPair) observer.LoggedEntry {
		return observer.LoggedEntry{
			Entry:   zapcore.Entry{Level: ErrorLevel, Message: _nonStringKeyErrMsg},
			Context: []Field{Array("invalid", invalidPairs(pairs))},
		}
	}
	ignoredError := func(err error) observer.LoggedEntry {
		return observer.LoggedEntry{
			Entry:   zapcore.Entry{Level: ErrorLevel, Message: _multipleErrMsg},
			Context: []Field{Error(err)},
		}
	}

	type withAny func(*SugaredLogger, ...interface{}) *SugaredLogger
	withMethods := []withAny{(*SugaredLogger).With, (*SugaredLogger).WithLazy}

	tests := []struct {
		desc     string
		args     []interface{}
		expected []Field
		errLogs  []observer.LoggedEntry
	}{
		{
			desc:     "nil args",
			args:     nil,
			expected: []Field{},
			errLogs:  nil,
		},
		{
			desc:     "empty slice of args",
			args:     []interface{}{},
			expected: []Field{},
			errLogs:  nil,
		},
		{
			desc:     "just a dangling key",
			args:     []interface{}{"should ignore"},
			expected: []Field{},
			errLogs:  []observer.LoggedEntry{ignored("should ignore")},
		},
		{
			desc:     "well-formed key-value pairs",
			args:     []interface{}{"foo", 42, "true", "bar"},
			expected: []Field{Int("foo", 42), String("true", "bar")},
			errLogs:  nil,
		},
		{
			desc:     "just a structured field",
			args:     []interface{}{Int("foo", 42)},
			expected: []Field{Int("foo", 42)},
			errLogs:  nil,
		},
		{
			desc:     "structured field and a dangling key",
			args:     []interface{}{Int("foo", 42), "dangling"},
			expected: []Field{Int("foo", 42)},
			errLogs:  []observer.LoggedEntry{ignored("dangling")},
		},
		{
			desc:     "structured field and a dangling non-string key",
			args:     []interface{}{Int("foo", 42), 13},
			expected: []Field{Int("foo", 42)},
			errLogs:  []observer.LoggedEntry{ignored(13)},
		},
		{
			desc:     "key-value pair and a dangling key",
			args:     []interface{}{"foo", 42, "dangling"},
			expected: []Field{Int("foo", 42)},
			errLogs:  []observer.LoggedEntry{ignored("dangling")},
		},
		{
			desc:     "pairs, a structured field, and a dangling key",
			args:     []interface{}{"first", "field", Int("foo", 42), "baz", "quux", "dangling"},
			expected: []Field{String("first", "field"), Int("foo", 42), String("baz", "quux")},
			errLogs:  []observer.LoggedEntry{ignored("dangling")},
		},
		{
			desc:     "one non-string key",
			args:     []interface{}{"foo", 42, true, "bar"},
			expected: []Field{Int("foo", 42)},
			errLogs:  []observer.LoggedEntry{nonString(invalidPair{2, true, "bar"})},
		},
		{
			desc:     "pairs, structured fields, non-string keys, and a dangling key",
			args:     []interface{}{"foo", 42, true, "bar", Int("structure", 11), 42, "reversed", "baz", "quux", "dangling"},
			expected: []Field{Int("foo", 42), Int("structure", 11), String("baz", "quux")},
			errLogs: []observer.LoggedEntry{
				ignored("dangling"),
				nonString(invalidPair{2, true, "bar"}, invalidPair{5, 42, "reversed"}),
			},
		},
		{
			desc:     "multiple errors",
			args:     []interface{}{errors.New("first"), errors.New("second"), errors.New("third")},
			expected: []Field{Error(errors.New("first"))},
			errLogs: []observer.LoggedEntry{
				ignoredError(errors.New("second")),
				ignoredError(errors.New("third")),
			},
		},
	}

	for _, tt := range tests {
		for _, withMethod := range withMethods {
			withSugar(t, DebugLevel, nil, func(logger *SugaredLogger, logs *observer.ObservedLogs) {
				withMethod(logger, tt.args...).Info("")
				output := logs.AllUntimed()
				if len(tt.errLogs) > 0 {
					for i := range tt.errLogs {
						assert.Equal(t, tt.errLogs[i], output[i], "Unexpected error log at position %d for scenario %s.", i, tt.desc)
					}
				}
				assert.Equal(t, len(tt.errLogs)+1, len(output), "Expected only one non-error message to be logged in scenario %s.", tt.desc)
				assert.Equal(t, tt.expected, output[len(tt.errLogs)].Context, "Unexpected message context in scenario %s.", tt.desc)
			})
		}
	}
}

func TestSugarWithCaptures(t *testing.T) {
	type withAny func(*SugaredLogger, ...interface{}) *SugaredLogger

	tests := []struct {
		name        string
		withMethods []withAny
		wantJSON    []string
	}{
		{
			name:        "with captures arguments at time of With",
			withMethods: []withAny{(*SugaredLogger).With},
			wantJSON: []string{
				`{
					"m": "hello 0",
					"a0": [0],
					"b0": [1]
				}`,
				`{
					"m": "world 0",
					"a0": [0],
					"c0": [2]
				}`,
			},
		},
		{
			name:        "lazy with captures arguments at time of Logging",
			withMethods: []withAny{(*SugaredLogger).WithLazy},
			wantJSON: []string{
				`{
					"m": "hello 0",
					"a0": [1],
					"b0": [1]
				}`,
				`{
					"m": "world 0",
					"a0": [1],
					"c0": [2]
				}`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			enc := zapcore.NewJSONEncoder(zapcore.EncoderConfig{
				MessageKey: "m",
			})

			var bs ztest.Buffer
			logger := New(zapcore.NewCore(enc, &bs, DebugLevel)).Sugar()

			for i, withMethod := range tt.withMethods {
				iStr := strconv.Itoa(i)
				x := 10 * i
				arr := zapcore.ArrayMarshalerFunc(func(enc zapcore.ArrayEncoder) error {
					enc.AppendInt(x)
					return nil
				})

				logger = withMethod(logger, Array("a"+iStr, arr))
				x++
				logger.Infow(fmt.Sprintf("hello %d", i), Array("b"+iStr, arr))
				x++
				logger = withMethod(logger, Array("c"+iStr, arr))
				logger.Infow(fmt.Sprintf("world %d", i))
			}

			if lines := bs.Lines(); assert.Len(t, lines, len(tt.wantJSON)) {
				for i, want := range tt.wantJSON {
					assert.JSONEq(t, want, lines[i], "Unexpected output from the %d'th log.", i)
				}
			}
		})
	}
}

func TestSugaredLoggerLevel(t *testing.T) {
	levels := []zapcore.Level{
		DebugLevel,
		InfoLevel,
		WarnLevel,
		ErrorLevel,
		DPanicLevel,
		PanicLevel,
		FatalLevel,
	}

	for _, lvl := range levels {
		lvl := lvl
		t.Run(lvl.String(), func(t *testing.T) {
			t.Parallel()

			core, _ := observer.New(lvl)
			log := New(core).Sugar()
			assert.Equal(t, lvl, log.Level())
		})
	}

	t.Run("Nop", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, zapcore.InvalidLevel, NewNop().Sugar().Level())
	})
}

func TestSugarFieldsInvalidPairs(t *testing.T) {
	withSugar(t, DebugLevel, nil, func(logger *SugaredLogger, logs *observer.ObservedLogs) {
		logger.With(42, "foo", []string{"bar"}, "baz").Info("")
		output := logs.AllUntimed()

		// Double-check that the actual message was logged.
		require.Equal(t, 2, len(output), "Unexpected number of entries logged.")
		require.Equal(t, observer.LoggedEntry{Context: []Field{}}, output[1], "Unexpected non-error log entry.")

		// Assert that the error message's structured fields serialize properly.
		require.Equal(t, 1, len(output[0].Context), "Expected one field in error entry context.")
		enc := zapcore.NewMapObjectEncoder()
		output[0].Context[0].AddTo(enc)
		assert.Equal(t, []interface{}{
			map[string]interface{}{"position": int64(0), "key": int64(42), "value": "foo"},
			map[string]interface{}{"position": int64(2), "key": []interface{}{"bar"}, "value": "baz"},
		}, enc.Fields["invalid"], "Unexpected output when logging invalid key-value pairs.")
	})
}

func TestSugarStructuredLogging(t *testing.T) {
	tests := []struct {
		msg       string
		expectMsg string
	}{
		{"foo", "foo"},
		{"", ""},
	}

	// Common to all test cases.
	var (
		err            = errors.New("qux")
		context        = []interface{}{"foo", "bar"}
		extra          = []interface{}{err, "baz", false}
		expectedFields = []Field{String("foo", "bar"), Error(err), Bool("baz", false)}
	)

	for _, tt := range tests {
		withSugar(t, DebugLevel, nil, func(logger *SugaredLogger, logs *observer.ObservedLogs) {
			logger.With(context...).Debugw(tt.msg, extra...)
			logger.With(context...).Infow(tt.msg, extra...)
			logger.With(context...).Warnw(tt.msg, extra...)
			logger.With(context...).Errorw(tt.msg, extra...)
			logger.With(context...).DPanicw(tt.msg, extra...)
			logger.With(context...).Logw(WarnLevel, tt.msg, extra...)

			expected := make([]observer.LoggedEntry, 6)
			for i, lvl := range []zapcore.Level{DebugLevel, InfoLevel, WarnLevel, ErrorLevel, DPanicLevel, WarnLevel} {
				expected[i] = observer.LoggedEntry{
					Entry:   zapcore.Entry{Message: tt.expectMsg, Level: lvl},
					Context: expectedFields,
				}
			}
			assert.Equal(t, expected, logs.AllUntimed(), "Unexpected log output.")
		})
	}
}

func TestSugarConcatenatingLogging(t *testing.T) {
	tests := []struct {
		args   []interface{}
		expect string
	}{
		{[]interface{}{nil}, "<nil>"},
	}

	// Common to all test cases.
	context := []interface{}{"foo", "bar"}
	expectedFields := []Field{String("foo", "bar")}

	for _, tt := range tests {
		withSugar(t, DebugLevel, nil, func(logger *SugaredLogger, logs *observer.ObservedLogs) {
			logger.With(context...).Debug(tt.args...)
			logger.With(context...).Info(tt.args...)
			logger.With(context...).Warn(tt.args...)
			logger.With(context...).Error(tt.args...)
			logger.With(context...).DPanic(tt.args...)
			logger.With(context...).Log(InfoLevel, tt.args...)

			expected := make([]observer.LoggedEntry, 6)
			for i, lvl := range []zapcore.Level{DebugLevel, InfoLevel, WarnLevel, ErrorLevel, DPanicLevel, InfoLevel} {
				expected[i] = observer.LoggedEntry{
					Entry:   zapcore.Entry{Message: tt.expect, Level: lvl},
					Context: expectedFields,
				}
			}
			assert.Equal(t, expected, logs.AllUntimed(), "Unexpected log output.")
		})
	}
}

func TestSugarTemplatedLogging(t *testing.T) {
	tests := []struct {
		format string
		args   []interface{}
		expect string
	}{
		{"", nil, ""},
		{"foo", nil, "foo"},
		// If the user fails to pass a template, degrade to fmt.Sprint.
		{"", []interface{}{"foo"}, "foo"},
	}

	// Common to all test cases.
	context := []interface{}{"foo", "bar"}
	expectedFields := []Field{String("foo", "bar")}

	for _, tt := range tests {
		withSugar(t, DebugLevel, nil, func(logger *SugaredLogger, logs *observer.ObservedLogs) {
			logger.With(context...).Debugf(tt.format, tt.args...)
			logger.With(context...).Infof(tt.format, tt.args...)
			logger.With(context...).Warnf(tt.format, tt.args...)
			logger.With(context...).Errorf(tt.format, tt.args...)
			logger.With(context...).DPanicf(tt.format, tt.args...)
			logger.With(context...).Logf(ErrorLevel, tt.format, tt.args...)

			expected := make([]observer.LoggedEntry, 6)
			for i, lvl := range []zapcore.Level{DebugLevel, InfoLevel, WarnLevel, ErrorLevel, DPanicLevel, ErrorLevel} {
				expected[i] = observer.LoggedEntry{
					Entry:   zapcore.Entry{Message: tt.expect, Level: lvl},
					Context: expectedFields,
				}
			}
			assert.Equal(t, expected, logs.AllUntimed(), "Unexpected log output.")
		})
	}
}

func TestSugarLnLogging(t *testing.T) {
	tests := []struct {
		args   []interface{}
		expect string
	}{
		{nil, ""},
		{[]interface{}{}, ""},
		{[]interface{}{""}, ""},
		{[]interface{}{"foo"}, "foo"},
		{[]interface{}{"foo", "bar"}, "foo bar"},
	}

	// Common to all test cases.
	context := []interface{}{"foo", "bar"}
	expectedFields := []Field{String("foo", "bar")}

	for _, tt := range tests {
		withSugar(t, DebugLevel, nil, func(logger *SugaredLogger, logs *observer.ObservedLogs) {
			logger.With(context...).Debugln(tt.args...)
			logger.With(context...).Infoln(tt.args...)
			logger.With(context...).Warnln(tt.args...)
			logger.With(context...).Errorln(tt.args...)
			logger.With(context...).DPanicln(tt.args...)
			logger.With(context...).Logln(InfoLevel, tt.args...)

			expected := make([]observer.LoggedEntry, 6)
			for i, lvl := range []zapcore.Level{DebugLevel, InfoLevel, WarnLevel, ErrorLevel, DPanicLevel, InfoLevel} {
				expected[i] = observer.LoggedEntry{
					Entry:   zapcore.Entry{Message: tt.expect, Level: lvl},
					Context: expectedFields,
				}
			}
			assert.Equal(t, expected, logs.AllUntimed(), "Unexpected log output.")
		})
	}
}

func TestSugarLnLoggingIgnored(t *testing.T) {
	withSugar(t, WarnLevel, nil, func(logger *SugaredLogger, logs *observer.ObservedLogs) {
		logger.Infoln("hello")
		assert.Zero(t, logs.Len(), "Expected zero log statements.")
	})
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
		{FatalLevel, func(s *SugaredLogger) { s.Panicw("foo") }, ""},
		{PanicLevel, func(s *SugaredLogger) { s.Panicw("foo") }, "foo"},
		{DebugLevel, func(s *SugaredLogger) { s.Panicw("foo") }, "foo"},
		{FatalLevel, func(s *SugaredLogger) { s.Panicln("foo") }, ""},
		{PanicLevel, func(s *SugaredLogger) { s.Panicln("foo") }, "foo"},
		{DebugLevel, func(s *SugaredLogger) { s.Panicln("foo") }, "foo"},
	}

	for _, tt := range tests {
		withSugar(t, tt.loggerLevel, nil, func(sugar *SugaredLogger, logs *observer.ObservedLogs) {
			assert.Panics(t, func() { tt.f(sugar) }, "Expected panic-level logger calls to panic.")
			if tt.expectedMsg != "" {
				assert.Equal(t, []observer.LoggedEntry{{
					Context: []Field{},
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
		{FatalLevel + 1, func(s *SugaredLogger) { s.Fatalw("foo") }, ""},
		{FatalLevel, func(s *SugaredLogger) { s.Fatalw("foo") }, "foo"},
		{DebugLevel, func(s *SugaredLogger) { s.Fatalw("foo") }, "foo"},
		{FatalLevel + 1, func(s *SugaredLogger) { s.Fatalln("foo") }, ""},
		{FatalLevel, func(s *SugaredLogger) { s.Fatalln("foo") }, "foo"},
		{DebugLevel, func(s *SugaredLogger) { s.Fatalln("foo") }, "foo"},
	}

	for _, tt := range tests {
		withSugar(t, tt.loggerLevel, nil, func(sugar *SugaredLogger, logs *observer.ObservedLogs) {
			stub := exit.WithStub(func() { tt.f(sugar) })
			assert.True(t, stub.Exited, "Expected all calls to fatal logger methods to exit process.")
			if tt.expectedMsg != "" {
				assert.Equal(t, []observer.LoggedEntry{{
					Context: []Field{},
					Entry:   zapcore.Entry{Message: tt.expectedMsg, Level: FatalLevel},
				}}, logs.AllUntimed(), "Unexpected log output.")
			} else {
				assert.Equal(t, 0, logs.Len(), "Didn't expect any log output.")
			}
		})
	}
}

func TestSugarAddCaller(t *testing.T) {
	tests := []struct {
		options []Option
		pat     string
	}{
		{opts(AddCaller()), `.+/sugar_test.go:[\d]+$`},
		{opts(AddCaller(), AddCallerSkip(1), AddCallerSkip(-1)), `.+/sugar_test.go:[\d]+$`},
		{opts(AddCaller(), AddCallerSkip(1)), `.+/common_test.go:[\d]+$`},
		{opts(AddCaller(), AddCallerSkip(1), AddCallerSkip(5)), `.+/src/runtime/.*:[\d]+$`},
	}
	for _, tt := range tests {
		withSugar(t, DebugLevel, tt.options, func(logger *SugaredLogger, logs *observer.ObservedLogs) {
			logger.Info("")
			output := logs.AllUntimed()
			assert.Equal(t, 1, len(output), "Unexpected number of logs written out.")
			assert.Regexp(
				t,
				tt.pat,
				output[0].Caller,
				"Expected to find package name and file name in output.",
			)
		})
	}
}

func TestSugarAddCallerFail(t *testing.T) {
	errBuf := &ztest.Buffer{}
	withSugar(t, DebugLevel, opts(AddCaller(), AddCallerSkip(1e3), ErrorOutput(errBuf)), func(log *SugaredLogger, logs *observer.ObservedLogs) {
		log.Info("Failure.")
		assert.Regexp(
			t,
			`Logger.check error: failed to get caller`,
			errBuf.String(),
			"Didn't find expected failure message.",
		)
		assert.Equal(
			t,
			logs.AllUntimed()[0].Message,
			"Failure.",
			"Expected original message to survive failures in runtime.Caller.")
	})
}

func TestSugarWithOptionsIncreaseLevel(t *testing.T) {
	withSugar(t, DebugLevel, nil, func(logger *SugaredLogger, logs *observer.ObservedLogs) {
		logger = logger.WithOptions(IncreaseLevel(WarnLevel))
		logger.Info("logger.Info")
		logger.Warn("logger.Warn")
		logger.Error("logger.Error")
		require.Equal(t, 2, logs.Len(), "expected only warn + error logs due to IncreaseLevel.")
		assert.Equal(
			t,
			logs.AllUntimed()[0].Message,
			"logger.Warn",
			"Expected first logged message to be warn level message",
		)
	})
}

func TestSugarLnWithOptionsIncreaseLevel(t *testing.T) {
	withSugar(t, DebugLevel, nil, func(logger *SugaredLogger, logs *observer.ObservedLogs) {
		logger = logger.WithOptions(IncreaseLevel(WarnLevel))
		logger.Infoln("logger.Infoln")
		logger.Warnln("logger.Warnln")
		logger.Errorln("logger.Errorln")
		require.Equal(t, 2, logs.Len(), "expected only warn + error logs due to IncreaseLevel.")
		assert.Equal(
			t,
			logs.AllUntimed()[0].Message,
			"logger.Warnln",
			"Expected first logged message to be warn level message",
		)
	})
}

func BenchmarkSugarSingleStrArg(b *testing.B) {
	withSugar(b, InfoLevel, nil /* opts* */, func(log *SugaredLogger, logs *observer.ObservedLogs) {
		for i := 0; i < b.N; i++ {
			log.Info("hello world")
		}
	})
}

func BenchmarkLnSugarSingleStrArg(b *testing.B) {
	withSugar(b, InfoLevel, nil /* opts* */, func(log *SugaredLogger, logs *observer.ObservedLogs) {
		for i := 0; i < b.N; i++ {
			log.Infoln("hello world")
		}
	})
}
