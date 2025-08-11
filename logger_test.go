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
	"sync"
	"sync/atomic"
	"testing"

	"go.uber.org/zap/internal/exit"
	"go.uber.org/zap/internal/ztest"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeCountingHook() (func(zapcore.Entry) error, *atomic.Int64) {
	count := &atomic.Int64{}
	h := func(zapcore.Entry) error {
		count.Add(1)
		return nil
	}
	return h, count
}

func TestLoggerAtomicLevel(t *testing.T) {
	// Test that the dynamic level applies to all ancestors and descendants.
	dl := NewAtomicLevel()

	withLogger(t, dl, nil, func(grandparent *Logger, _ *observer.ObservedLogs) {
		parent := grandparent.With(Int("generation", 1))
		child := parent.With(Int("generation", 2))

		tests := []struct {
			setLevel  zapcore.Level
			testLevel zapcore.Level
			enabled   bool
		}{
			{DebugLevel, DebugLevel, true},
			{InfoLevel, DebugLevel, false},
			{WarnLevel, PanicLevel, true},
		}

		for _, tt := range tests {
			dl.SetLevel(tt.setLevel)
			for _, logger := range []*Logger{grandparent, parent, child} {
				if tt.enabled {
					assert.NotNil(
						t,
						logger.Check(tt.testLevel, ""),
						"Expected level %s to be enabled after setting level %s.", tt.testLevel, tt.setLevel,
					)
				} else {
					assert.Nil(
						t,
						logger.Check(tt.testLevel, ""),
						"Expected level %s to be enabled after setting level %s.", tt.testLevel, tt.setLevel,
					)
				}
			}
		}
	})
}

func TestLoggerLevel(t *testing.T) {
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
			log := New(core)
			assert.Equal(t, lvl, log.Level())
		})
	}

	t.Run("Nop", func(t *testing.T) {
		assert.Equal(t, zapcore.InvalidLevel, NewNop().Level())
	})
}

func TestLoggerInitialFields(t *testing.T) {
	fieldOpts := opts(Fields(Int("foo", 42), String("bar", "baz")))
	withLogger(t, DebugLevel, fieldOpts, func(logger *Logger, logs *observer.ObservedLogs) {
		logger.Info("")
		assert.Equal(
			t,
			observer.LoggedEntry{Context: []Field{Int("foo", 42), String("bar", "baz")}},
			logs.AllUntimed()[0],
			"Unexpected output with initial fields set.",
		)
	})
}

func TestLoggerWith(t *testing.T) {
	tests := []struct {
		name          string
		initialFields []Field
		withMethod    func(*Logger, ...Field) *Logger
	}{
		{
			"regular non lazy logger",
			[]Field{Int("foo", 42)},
			(*Logger).With,
		},
		{
			"regular non lazy logger no initial fields",
			[]Field{},
			(*Logger).With,
		},
		{
			"lazy with logger",
			[]Field{Int("foo", 42)},
			(*Logger).WithLazy,
		},
		{
			"lazy with logger no initial fields",
			[]Field{},
			(*Logger).WithLazy,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withLogger(t, DebugLevel, opts(Fields(tt.initialFields...)), func(logger *Logger, logs *observer.ObservedLogs) {
				// Child loggers should have copy-on-write semantics, so two children
				// shouldn't stomp on each other's fields or affect the parent's fields.
				tt.withMethod(logger).Info("")
				tt.withMethod(logger, String("one", "two")).Info("")
				tt.withMethod(logger, String("three", "four")).Info("")
				tt.withMethod(logger, String("five", "six")).With(String("seven", "eight")).Info("")
				logger.Info("")

				assert.Equal(t, []observer.LoggedEntry{
					{Context: tt.initialFields},
					{Context: append(tt.initialFields, String("one", "two"))},
					{Context: append(tt.initialFields, String("three", "four"))},
					{Context: append(tt.initialFields, String("five", "six"), String("seven", "eight"))},
					{Context: tt.initialFields},
				}, logs.AllUntimed(), "Unexpected cross-talk between child loggers.")
			})
		})
	}
}

func TestLoggerWithCaptures(t *testing.T) {
	type withF func(*Logger, ...Field) *Logger
	tests := []struct {
		name        string
		withMethods []withF
		wantJSON    []string
	}{
		{
			name:        "regular with captures arguments at time of With",
			withMethods: []withF{(*Logger).With},
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
			name:        "lazy with captures arguments at time of With or Logging",
			withMethods: []withF{(*Logger).WithLazy},
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
		{
			name:        "2x With captures arguments at time of each With",
			withMethods: []withF{(*Logger).With, (*Logger).With},
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
				`{
					"m": "hello 1",
					"a0": [0],
					"c0": [2],
					"a1": [10],
					"b1": [11]
				}`,
				`{
					"m": "world 1",
					"a0": [0],
					"c0": [2],
					"a1": [10],
					"c1": [12]
				}`,
			},
		},
		{
			name:        "2x WithLazy. Captures arguments only at logging time.",
			withMethods: []withF{(*Logger).WithLazy, (*Logger).WithLazy},
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
				`{
					"m": "hello 1",
					"a0": [1],
					"c0": [2],
					"a1": [11],
					"b1": [11]
				}`,
				`{
					"m": "world 1",
					"a0": [1],
					"c0": [2],
					"a1": [11],
					"c1": [12]
				}`,
			},
		},
		{
			name:        "WithLazy then With",
			withMethods: []withF{(*Logger).WithLazy, (*Logger).With},
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
				`{
					"m": "hello 1",
					"a0": [1],
					"c0": [2],
					"a1": [10],
					"b1": [11]
				}`,
				`{
					"m": "world 1",
					"a0": [1],
					"c0": [2],
					"a1": [10],
					"c1": [12]
				}`,
			},
		},
		{
			name:        "With then WithLazy",
			withMethods: []withF{(*Logger).With, (*Logger).WithLazy},
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
				`{
					"m": "hello 1",
					"a0": [0],
					"c0": [2],
					"a1": [11],
					"b1": [11]
				}`,
				`{
					"m": "world 1",
					"a0": [0],
					"c0": [2],
					"a1": [11],
					"c1": [12]
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
			logger := New(zapcore.NewCore(enc, &bs, DebugLevel))

			for i, withMethod := range tt.withMethods {

				iStr := strconv.Itoa(i)
				x := 10 * i
				arr := zapcore.ArrayMarshalerFunc(func(enc zapcore.ArrayEncoder) error {
					enc.AppendInt(x)
					return nil
				})

				// Demonstrate the arguments are captured when With() and Info() are invoked.
				logger = withMethod(logger, Array("a"+iStr, arr))
				x++
				logger.Info(fmt.Sprintf("hello %d", i), Array("b"+iStr, arr))
				x++
				logger = withMethod(logger, Array("c"+iStr, arr))
				logger.Info(fmt.Sprintf("world %d", i))
			}

			if lines := bs.Lines(); assert.Len(t, lines, len(tt.wantJSON)) {
				for i, want := range tt.wantJSON {
					assert.JSONEq(t, want, lines[i], "Unexpected output from the %d'th log.", i)
				}
			}
		})
	}
}

func TestLoggerLogPanic(t *testing.T) {
	for _, tt := range []struct {
		do       func(*Logger)
		should   bool
		expected string
	}{
		{func(logger *Logger) { logger.Check(PanicLevel, "foo").Write() }, true, "foo"},
		{func(logger *Logger) { logger.Log(PanicLevel, "bar") }, true, "bar"},
		{func(logger *Logger) { logger.Panic("baz") }, true, "baz"},
	} {
		withLogger(t, DebugLevel, nil, func(logger *Logger, logs *observer.ObservedLogs) {
			if tt.should {
				assert.Panics(t, func() { tt.do(logger) }, "Expected panic")
			} else {
				assert.NotPanics(t, func() { tt.do(logger) }, "Expected no panic")
			}

			output := logs.AllUntimed()
			assert.Equal(t, 1, len(output), "Unexpected number of logs.")
			assert.Equal(t, 0, len(output[0].Context), "Unexpected context on first log.")
			assert.Equal(
				t,
				zapcore.Entry{Message: tt.expected, Level: PanicLevel},
				output[0].Entry,
				"Unexpected output from panic-level Log.",
			)
		})
	}
}

func TestLoggerLogFatal(t *testing.T) {
	for _, tt := range []struct {
		do       func(*Logger)
		expected string
	}{
		{func(logger *Logger) { logger.Check(FatalLevel, "foo").Write() }, "foo"},
		{func(logger *Logger) { logger.Log(FatalLevel, "bar") }, "bar"},
		{func(logger *Logger) { logger.Fatal("baz") }, "baz"},
	} {
		withLogger(t, DebugLevel, nil, func(logger *Logger, logs *observer.ObservedLogs) {
			stub := exit.WithStub(func() {
				tt.do(logger)
			})
			assert.True(t, stub.Exited, "Expected Fatal logger call to terminate process.")
			output := logs.AllUntimed()
			assert.Equal(t, 1, len(output), "Unexpected number of logs.")
			assert.Equal(t, 0, len(output[0].Context), "Unexpected context on first log.")
			assert.Equal(
				t,
				zapcore.Entry{Message: tt.expected, Level: FatalLevel},
				output[0].Entry,
				"Unexpected output from fatal-level Log.",
			)
		})
	}
}

func TestLoggerLeveledMethods(t *testing.T) {
	withLogger(t, DebugLevel, nil, func(logger *Logger, logs *observer.ObservedLogs) {
		tests := []struct {
			method        func(string, ...Field)
			expectedLevel zapcore.Level
		}{
			{logger.Debug, DebugLevel},
			{logger.Info, InfoLevel},
			{logger.Warn, WarnLevel},
			{logger.Error, ErrorLevel},
			{logger.DPanic, DPanicLevel},
		}
		for i, tt := range tests {
			tt.method("")
			output := logs.AllUntimed()
			assert.Equal(t, i+1, len(output), "Unexpected number of logs.")
			assert.Equal(t, 0, len(output[i].Context), "Unexpected context on first log.")
			assert.Equal(
				t,
				zapcore.Entry{Level: tt.expectedLevel},
				output[i].Entry,
				"Unexpected output from %s-level logger method.", tt.expectedLevel)
		}
	})
}

func TestLoggerLogLevels(t *testing.T) {
	withLogger(t, DebugLevel, nil, func(logger *Logger, logs *observer.ObservedLogs) {
		levels := []zapcore.Level{
			DebugLevel,
			InfoLevel,
			WarnLevel,
			ErrorLevel,
			DPanicLevel,
		}
		for i, level := range levels {
			logger.Log(level, "")
			output := logs.AllUntimed()
			assert.Equal(t, i+1, len(output), "Unexpected number of logs.")
			assert.Equal(t, 0, len(output[i].Context), "Unexpected context on first log.")
			assert.Equal(
				t,
				zapcore.Entry{Level: level},
				output[i].Entry,
				"Unexpected output from %s-level logger method.", level)
		}
	})
}

func TestLoggerAlwaysPanics(t *testing.T) {
	// Users can disable writing out panic-level logs, but calls to logger.Panic()
	// should still call panic().
	withLogger(t, FatalLevel, nil, func(logger *Logger, logs *observer.ObservedLogs) {
		msg := "Even if output is disabled, logger.Panic should always panic."
		assert.Panics(t, func() { logger.Panic("foo") }, msg)
		assert.Panics(t, func() { logger.Log(PanicLevel, "foo") }, msg)
		assert.Panics(t, func() {
			if ce := logger.Check(PanicLevel, "foo"); ce != nil {
				ce.Write()
			}
		}, msg)
		assert.Equal(t, 0, logs.Len(), "Panics shouldn't be written out if PanicLevel is disabled.")
	})
}

func TestLoggerAlwaysFatals(t *testing.T) {
	// Users can disable writing out fatal-level logs, but calls to logger.Fatal()
	// should still terminate the process.
	withLogger(t, FatalLevel+1, nil, func(logger *Logger, logs *observer.ObservedLogs) {
		stub := exit.WithStub(func() { logger.Fatal("") })
		assert.True(t, stub.Exited, "Expected calls to logger.Fatal to terminate process.")

		stub = exit.WithStub(func() { logger.Log(FatalLevel, "") })
		assert.True(t, stub.Exited, "Expected calls to logger.Fatal to terminate process.")

		stub = exit.WithStub(func() {
			if ce := logger.Check(FatalLevel, ""); ce != nil {
				ce.Write()
			}
		})
		assert.True(t, stub.Exited, "Expected calls to logger.Check(FatalLevel, ...) to terminate process.")

		assert.Equal(t, 0, logs.Len(), "Shouldn't write out logs when fatal-level logging is disabled.")
	})
}

func TestLoggerDPanic(t *testing.T) {
	withLogger(t, DebugLevel, nil, func(logger *Logger, logs *observer.ObservedLogs) {
		assert.NotPanics(t, func() { logger.DPanic("") })
		assert.Equal(
			t,
			[]observer.LoggedEntry{{Entry: zapcore.Entry{Level: DPanicLevel}, Context: []Field{}}},
			logs.AllUntimed(),
			"Unexpected log output from DPanic in production mode.",
		)
	})
	withLogger(t, DebugLevel, opts(Development()), func(logger *Logger, logs *observer.ObservedLogs) {
		assert.Panics(t, func() { logger.DPanic("") })
		assert.Equal(
			t,
			[]observer.LoggedEntry{{Entry: zapcore.Entry{Level: DPanicLevel}, Context: []Field{}}},
			logs.AllUntimed(),
			"Unexpected log output from DPanic in development mode.",
		)
	})
}

func TestLoggerNoOpsDisabledLevels(t *testing.T) {
	withLogger(t, WarnLevel, nil, func(logger *Logger, logs *observer.ObservedLogs) {
		logger.Info("silence!")
		assert.Equal(
			t,
			[]observer.LoggedEntry{},
			logs.AllUntimed(),
			"Expected logging at a disabled level to produce no output.",
		)
	})
}

func TestLoggerNames(t *testing.T) {
	tests := []struct {
		names    []string
		expected string
	}{
		{nil, ""},
		{[]string{""}, ""},
		{[]string{"foo"}, "foo"},
		{[]string{"foo", ""}, "foo"},
		{[]string{"foo", "bar"}, "foo.bar"},
		{[]string{"foo.bar", "baz"}, "foo.bar.baz"},
		// Garbage in, garbage out.
		{[]string{"foo.", "bar"}, "foo..bar"},
		{[]string{"foo", ".bar"}, "foo..bar"},
		{[]string{"foo.", ".bar"}, "foo...bar"},
	}

	for _, tt := range tests {
		withLogger(t, DebugLevel, nil, func(log *Logger, logs *observer.ObservedLogs) {
			for _, n := range tt.names {
				log = log.Named(n)
			}
			log.Info("")
			require.Equal(t, 1, logs.Len(), "Expected only one log entry to be written.")
			assert.Equal(t, tt.expected, logs.AllUntimed()[0].LoggerName, "Unexpected logger name from entry.")
			assert.Equal(t, tt.expected, log.Name(), "Unexpected logger name.")
		})
		withSugar(t, DebugLevel, nil, func(log *SugaredLogger, logs *observer.ObservedLogs) {
			for _, n := range tt.names {
				log = log.Named(n)
			}
			log.Infow("")
			require.Equal(t, 1, logs.Len(), "Expected only one log entry to be written.")
			assert.Equal(t, tt.expected, logs.AllUntimed()[0].LoggerName, "Unexpected logger name from entry.")
			assert.Equal(t, tt.expected, log.base.Name(), "Unexpected logger name.")
		})
	}
}

func TestLoggerWriteFailure(t *testing.T) {
	errSink := &ztest.Buffer{}
	logger := New(
		zapcore.NewCore(
			zapcore.NewJSONEncoder(NewProductionConfig().EncoderConfig),
			zapcore.Lock(zapcore.AddSync(ztest.FailWriter{})),
			DebugLevel,
		),
		ErrorOutput(errSink),
	)

	logger.Info("foo")
	// Should log the error.
	assert.Regexp(t, `write error: failed`, errSink.Stripped(), "Expected to log the error to the error output.")
	assert.True(t, errSink.Called(), "Expected logging an internal error to call Sync the error sink.")
}

func TestLoggerSync(t *testing.T) {
	withLogger(t, DebugLevel, nil, func(logger *Logger, _ *observer.ObservedLogs) {
		assert.NoError(t, logger.Sync(), "Expected syncing a test logger to succeed.")
		assert.NoError(t, logger.Sugar().Sync(), "Expected syncing a sugared logger to succeed.")
	})
}

func TestLoggerSyncFail(t *testing.T) {
	noSync := &ztest.Buffer{}
	err := errors.New("fail")
	noSync.SetError(err)
	logger := New(zapcore.NewCore(
		zapcore.NewJSONEncoder(zapcore.EncoderConfig{}),
		noSync,
		DebugLevel,
	))
	assert.Equal(t, err, logger.Sync(), "Expected Logger.Sync to propagate errors.")
	assert.Equal(t, err, logger.Sugar().Sync(), "Expected SugaredLogger.Sync to propagate errors.")
}

func TestLoggerAddCaller(t *testing.T) {
	tests := []struct {
		options []Option
		pat     string
	}{
		{opts(), `^undefined$`},
		{opts(WithCaller(false)), `^undefined$`},
		{opts(AddCaller()), `.+/logger_test.go:[\d]+$`},
		{opts(AddCaller(), WithCaller(false)), `^undefined$`},
		{opts(WithCaller(true)), `.+/logger_test.go:[\d]+$`},
		{opts(WithCaller(true), WithCaller(false)), `^undefined$`},
		{opts(AddCaller(), AddCallerSkip(1), AddCallerSkip(-1)), `.+/logger_test.go:[\d]+$`},
		{opts(AddCaller(), AddCallerSkip(1)), `.+/common_test.go:[\d]+$`},
		{opts(AddCaller(), AddCallerSkip(1), AddCallerSkip(3)), `.+/src/runtime/.*:[\d]+$`},
	}
	for _, tt := range tests {
		withLogger(t, DebugLevel, tt.options, func(logger *Logger, logs *observer.ObservedLogs) {
			// Make sure that sugaring and desugaring resets caller skip properly.
			logger = logger.Sugar().Desugar()
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

func TestLoggerAddCallerFunction(t *testing.T) {
	tests := []struct {
		options         []Option
		loggerFunction  string
		sugaredFunction string
	}{
		{
			options:         opts(),
			loggerFunction:  "",
			sugaredFunction: "",
		},
		{
			options:         opts(WithCaller(false)),
			loggerFunction:  "",
			sugaredFunction: "",
		},
		{
			options:         opts(AddCaller()),
			loggerFunction:  "go.uber.org/zap.infoLog",
			sugaredFunction: "go.uber.org/zap.infoLogSugared",
		},
		{
			options:         opts(AddCaller(), WithCaller(false)),
			loggerFunction:  "",
			sugaredFunction: "",
		},
		{
			options:         opts(WithCaller(true)),
			loggerFunction:  "go.uber.org/zap.infoLog",
			sugaredFunction: "go.uber.org/zap.infoLogSugared",
		},
		{
			options:         opts(WithCaller(true), WithCaller(false)),
			loggerFunction:  "",
			sugaredFunction: "",
		},
		{
			options:         opts(AddCaller(), AddCallerSkip(1), AddCallerSkip(-1)),
			loggerFunction:  "go.uber.org/zap.infoLog",
			sugaredFunction: "go.uber.org/zap.infoLogSugared",
		},
		{
			options:         opts(AddCaller(), AddCallerSkip(2)),
			loggerFunction:  "go.uber.org/zap.withLogger",
			sugaredFunction: "go.uber.org/zap.withLogger",
		},
		{
			options:         opts(AddCaller(), AddCallerSkip(2), AddCallerSkip(3)),
			loggerFunction:  "runtime.goexit",
			sugaredFunction: "runtime.goexit",
		},
	}
	for _, tt := range tests {
		withLogger(t, DebugLevel, tt.options, func(logger *Logger, logs *observer.ObservedLogs) {
			// Make sure that sugaring and desugaring resets caller skip properly.
			logger = logger.Sugar().Desugar()
			infoLog(logger, "")
			infoLogSugared(logger.Sugar(), "")
			infoLog(logger.Sugar().Desugar(), "")

			entries := logs.AllUntimed()
			assert.Equal(t, 3, len(entries), "Unexpected number of logs written out.")
			for _, entry := range []observer.LoggedEntry{entries[0], entries[2]} {
				assert.Regexp(
					t,
					tt.loggerFunction,
					entry.Caller.Function,
					"Expected to find function name in output.",
				)
			}
			assert.Regexp(
				t,
				tt.sugaredFunction,
				entries[1].Caller.Function,
				"Expected to find function name in output.",
			)
		})
	}
}

func TestLoggerAddCallerFail(t *testing.T) {
	errBuf := &ztest.Buffer{}
	withLogger(t, DebugLevel, opts(AddCaller(), AddCallerSkip(1e3), ErrorOutput(errBuf)), func(log *Logger, logs *observer.ObservedLogs) {
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
		assert.Equal(
			t,
			logs.AllUntimed()[0].Caller.Function,
			"",
			"Expected function name to be empty string.")
	})
}

func TestLoggerReplaceCore(t *testing.T) {
	replace := WrapCore(func(zapcore.Core) zapcore.Core {
		return zapcore.NewNopCore()
	})
	withLogger(t, DebugLevel, opts(replace), func(logger *Logger, logs *observer.ObservedLogs) {
		logger.Debug("")
		logger.Info("")
		logger.Warn("")
		assert.Equal(t, 0, logs.Len(), "Expected no-op core to write no logs.")
	})
}

func TestLoggerIncreaseLevel(t *testing.T) {
	withLogger(t, DebugLevel, opts(IncreaseLevel(WarnLevel)), func(logger *Logger, logs *observer.ObservedLogs) {
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

func TestLoggerHooks(t *testing.T) {
	hook, seen := makeCountingHook()
	withLogger(t, DebugLevel, opts(Hooks(hook)), func(logger *Logger, logs *observer.ObservedLogs) {
		logger.Debug("")
		logger.Info("")
	})
	assert.Equal(t, int64(2), seen.Load(), "Hook saw an unexpected number of logs.")
}

func TestLoggerConcurrent(t *testing.T) {
	withLogger(t, DebugLevel, nil, func(logger *Logger, logs *observer.ObservedLogs) {
		child := logger.With(String("foo", "bar"))

		wg := &sync.WaitGroup{}
		runConcurrently(5, 10, wg, func() {
			logger.Info("", String("foo", "bar"))
		})
		runConcurrently(5, 10, wg, func() {
			child.Info("")
		})

		wg.Wait()

		// Make sure the output doesn't contain interspersed entries.
		assert.Equal(t, 100, logs.Len(), "Unexpected number of logs written out.")
		for _, obs := range logs.AllUntimed() {
			assert.Equal(
				t,
				observer.LoggedEntry{
					Entry:   zapcore.Entry{Level: InfoLevel},
					Context: []Field{String("foo", "bar")},
				},
				obs,
				"Unexpected log output.",
			)
		}
	})
}

func TestLoggerFatalOnNoop(t *testing.T) {
	exitStub := exit.Stub()
	defer exitStub.Unstub()
	core, _ := observer.New(InfoLevel)

	// We don't allow a no-op fatal hook.
	New(core, WithFatalHook(zapcore.WriteThenNoop)).Fatal("great sadness")
	assert.True(t, exitStub.Exited, "must exit for WriteThenNoop")
	assert.Equal(t, 1, exitStub.Code, "must exit with status 1 for WriteThenNoop")
}

func TestLoggerCustomOnPanic(t *testing.T) {
	tests := []struct {
		msg          string
		level        zapcore.Level
		opts         []Option
		finished     bool
		want         []observer.LoggedEntry
		recoverValue any
	}{
		{
			msg:      "panic with nil hook",
			level:    PanicLevel,
			opts:     opts(WithPanicHook(nil)),
			finished: false,
			want: []observer.LoggedEntry{
				{
					Entry:   zapcore.Entry{Level: PanicLevel, Message: "foobar"},
					Context: []Field{},
				},
			},
			recoverValue: "foobar",
		},
		{
			msg:      "panic with noop hook",
			level:    PanicLevel,
			opts:     opts(WithPanicHook(zapcore.WriteThenNoop)),
			finished: false,
			want: []observer.LoggedEntry{
				{
					Entry:   zapcore.Entry{Level: PanicLevel, Message: "foobar"},
					Context: []Field{},
				},
			},
			recoverValue: "foobar",
		},
		{
			msg:      "no panic with goexit hook",
			level:    PanicLevel,
			opts:     opts(WithPanicHook(zapcore.WriteThenGoexit)),
			finished: false,
			want: []observer.LoggedEntry{
				{
					Entry:   zapcore.Entry{Level: PanicLevel, Message: "foobar"},
					Context: []Field{},
				},
			},
			recoverValue: nil,
		},
		{
			msg:      "dpanic no panic in development mode with goexit hook",
			level:    DPanicLevel,
			opts:     opts(WithPanicHook(zapcore.WriteThenGoexit), Development()),
			finished: false,
			want: []observer.LoggedEntry{
				{
					Entry:   zapcore.Entry{Level: DPanicLevel, Message: "foobar"},
					Context: []Field{},
				},
			},
			recoverValue: nil,
		},
		{
			msg:      "dpanic panic in development mode with noop hook",
			level:    DPanicLevel,
			opts:     opts(WithPanicHook(zapcore.WriteThenNoop), Development()),
			finished: false,
			want: []observer.LoggedEntry{
				{
					Entry:   zapcore.Entry{Level: DPanicLevel, Message: "foobar"},
					Context: []Field{},
				},
			},
			recoverValue: "foobar",
		},
		{
			msg:      "dpanic no exit in production mode with goexit hook",
			level:    DPanicLevel,
			opts:     opts(WithPanicHook(zapcore.WriteThenPanic)),
			finished: true,
			want: []observer.LoggedEntry{
				{
					Entry:   zapcore.Entry{Level: DPanicLevel, Message: "foobar"},
					Context: []Field{},
				},
			},
			recoverValue: nil,
		},
		{
			msg:      "dpanic no panic in production mode with panic hook",
			level:    DPanicLevel,
			opts:     opts(WithPanicHook(zapcore.WriteThenPanic)),
			finished: true,
			want: []observer.LoggedEntry{
				{
					Entry:   zapcore.Entry{Level: DPanicLevel, Message: "foobar"},
					Context: []Field{},
				},
			},
			recoverValue: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.msg, func(t *testing.T) {
			withLogger(t, InfoLevel, tt.opts, func(logger *Logger, logs *observer.ObservedLogs) {
				var finished bool
				recovered := make(chan any)
				go func() {
					defer func() {
						recovered <- recover()
					}()

					logger.Log(tt.level, "foobar")
					finished = true
				}()

				assert.Equal(t, tt.recoverValue, <-recovered, "unexpected value from recover()")
				assert.Equal(t, tt.finished, finished, "expect goroutine finished state doesn't match")
				assert.Equal(t, tt.want, logs.AllUntimed(), "unexpected logs")
			})
		})
	}
}

func TestLoggerCustomOnFatal(t *testing.T) {
	tests := []struct {
		msg          string
		onFatal      zapcore.CheckWriteAction
		recoverValue interface{}
	}{
		{
			msg:          "panic",
			onFatal:      zapcore.WriteThenPanic,
			recoverValue: "fatal",
		},
		{
			msg:          "goexit",
			onFatal:      zapcore.WriteThenGoexit,
			recoverValue: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.msg, func(t *testing.T) {
			withLogger(t, InfoLevel, opts(OnFatal(tt.onFatal)), func(logger *Logger, logs *observer.ObservedLogs) {
				var finished bool
				recovered := make(chan interface{})
				go func() {
					defer func() {
						recovered <- recover()
					}()

					logger.Fatal("fatal")
					finished = true
				}()

				assert.Equal(t, tt.recoverValue, <-recovered, "unexpected value from recover()")
				assert.False(t, finished, "expect goroutine to not finish after Fatal")

				assert.Equal(t, []observer.LoggedEntry{{
					Entry:   zapcore.Entry{Level: FatalLevel, Message: "fatal"},
					Context: []Field{},
				}}, logs.AllUntimed(), "unexpected logs")
			})
		})
	}
}

type customWriteHook struct {
	called bool
}

func (h *customWriteHook) OnWrite(_ *zapcore.CheckedEntry, _ []Field) {
	h.called = true
}

func TestLoggerWithFatalHook(t *testing.T) {
	var h customWriteHook
	withLogger(t, InfoLevel, opts(WithFatalHook(&h)), func(logger *Logger, logs *observer.ObservedLogs) {
		logger.Fatal("great sadness")
		assert.True(t, h.called)
		assert.Equal(t, 1, logs.FilterLevelExact(FatalLevel).Len())
	})
}

func TestNopLogger(t *testing.T) {
	logger := NewNop()

	t.Run("basic levels", func(t *testing.T) {
		logger.Debug("foo", String("k", "v"))
		logger.Info("bar", Int("x", 42))
		logger.Warn("baz", Strings("ks", []string{"a", "b"}))
		logger.Error("qux", Error(errors.New("great sadness")))
	})

	t.Run("DPanic", func(t *testing.T) {
		logger.With(String("component", "whatever")).DPanic("stuff")
	})

	t.Run("Panic", func(t *testing.T) {
		assert.Panics(t, func() {
			logger.Panic("great sadness")
		}, "Nop logger should still cause panics.")
	})
}

func TestMust(t *testing.T) {
	t.Run("must without an error does not panic", func(t *testing.T) {
		assert.NotPanics(t, func() { Must(NewNop(), nil) }, "must paniced with no error")
	})

	t.Run("must with an error panics", func(t *testing.T) {
		assert.Panics(t, func() { Must(nil, errors.New("an error")) }, "must did not panic with an error")
	})
}

func infoLog(logger *Logger, msg string, fields ...Field) {
	logger.Info(msg, fields...)
}

func infoLogSugared(logger *SugaredLogger, args ...interface{}) {
	logger.Info(args...)
}
