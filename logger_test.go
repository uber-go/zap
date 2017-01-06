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
	"fmt"
	"io/ioutil"
	"strings"
	"sync"
	"testing"

	"go.uber.org/zap/spywrite"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDynamicLevel(t *testing.T) {
	lvl := DynamicLevel()
	assert.Equal(t, InfoLevel, lvl.Level(), "Unexpected initial level.")
	lvl.SetLevel(ErrorLevel)
	assert.Equal(t, ErrorLevel, lvl.Level(), "Unexpected level after SetLevel.")
}

func TestDynamicLevel_concurrentMutation(t *testing.T) {
	lvl := DynamicLevel()
	// Trigger races for non-atomic level mutations.
	proceed := make(chan struct{})
	wg := &sync.WaitGroup{}
	runConcurrently(10, 100, wg, func() {
		<-proceed
		lvl.Level()
	})
	runConcurrently(10, 100, wg, func() {
		<-proceed
		lvl.SetLevel(WarnLevel)
	})
	close(proceed)
	wg.Wait()
}

func TestJSONLogger_DynamicLevel(t *testing.T) {
	// Test that the DynamicLevel applys to all ancestors and descendants.
	dl := DynamicLevel()
	withJSONLogger(t, dl, nil, func(grandparent Logger, buf *testBuffer) {
		parent := grandparent.With(Int("generation", 2))
		child := parent.With(Int("generation", 3))
		all := []Logger{grandparent, parent, child}

		assert.Equal(t, InfoLevel, dl.Level(), "expected initial InfoLevel")

		for round, lvl := range []Level{InfoLevel, DebugLevel, WarnLevel} {
			dl.SetLevel(lvl)
			for loggerI, log := range all {
				log.Debug("@debug", Int("round", round), Int("logger", loggerI))
			}
			for loggerI, log := range all {
				log.Info("@info", Int("round", round), Int("logger", loggerI))
			}
			for loggerI, log := range all {
				log.Warn("@warn", Int("round", round), Int("logger", loggerI))
			}
		}

		assert.Equal(t, []string{
			// round 0 at InfoLevel
			`{"level":"info","msg":"@info","round":0,"logger":0}`,
			`{"level":"info","msg":"@info","generation":2,"round":0,"logger":1}`,
			`{"level":"info","msg":"@info","generation":2,"generation":3,"round":0,"logger":2}`,
			`{"level":"warn","msg":"@warn","round":0,"logger":0}`,
			`{"level":"warn","msg":"@warn","generation":2,"round":0,"logger":1}`,
			`{"level":"warn","msg":"@warn","generation":2,"generation":3,"round":0,"logger":2}`,

			// round 1 at DebugLevel
			`{"level":"debug","msg":"@debug","round":1,"logger":0}`,
			`{"level":"debug","msg":"@debug","generation":2,"round":1,"logger":1}`,
			`{"level":"debug","msg":"@debug","generation":2,"generation":3,"round":1,"logger":2}`,
			`{"level":"info","msg":"@info","round":1,"logger":0}`,
			`{"level":"info","msg":"@info","generation":2,"round":1,"logger":1}`,
			`{"level":"info","msg":"@info","generation":2,"generation":3,"round":1,"logger":2}`,
			`{"level":"warn","msg":"@warn","round":1,"logger":0}`,
			`{"level":"warn","msg":"@warn","generation":2,"round":1,"logger":1}`,
			`{"level":"warn","msg":"@warn","generation":2,"generation":3,"round":1,"logger":2}`,

			// round 2 at WarnLevel
			`{"level":"warn","msg":"@warn","round":2,"logger":0}`,
			`{"level":"warn","msg":"@warn","generation":2,"round":2,"logger":1}`,
			`{"level":"warn","msg":"@warn","generation":2,"generation":3,"round":2,"logger":2}`,
		}, strings.Split(buf.Stripped(), "\n"))
	})
}

func TestJSONLoggerInitialFields(t *testing.T) {
	fieldOpts := opts(Fields(Int("foo", 42), String("bar", "baz")))
	withJSONLogger(t, DebugLevel, fieldOpts, func(logger Logger, buf *testBuffer) {
		logger.Info("")
		assert.Equal(t,
			`{"level":"info","msg":"","foo":42,"bar":"baz"}`,
			buf.Stripped(),
			"Unexpected output with initial fields set.",
		)
	})
}

func TestJSONLoggerWith(t *testing.T) {
	fieldOpts := opts(Fields(Int("foo", 42)))
	withJSONLogger(t, DebugLevel, fieldOpts, func(logger Logger, buf *testBuffer) {
		// Child loggers should have copy-on-write semantics, so two children
		// shouldn't stomp on each other's fields or affect the parent's fields.
		logger.With(String("one", "two")).Debug("")
		logger.With(String("three", "four")).Debug("")
		logger.Debug("")
		assert.Equal(t, []string{
			`{"level":"debug","msg":"","foo":42,"one":"two"}`,
			`{"level":"debug","msg":"","foo":42,"three":"four"}`,
			`{"level":"debug","msg":"","foo":42}`,
		}, buf.Lines(), "Unexpected cross-talk between child loggers.")
	})
}

func TestJSONLoggerLogPanic(t *testing.T) {
	for _, tc := range []struct {
		do       func(Logger)
		should   bool
		expected string
	}{
		{func(logger Logger) { logger.Check(PanicLevel, "bar").Write() }, true, `{"level":"panic","msg":"bar"}`},
		{func(logger Logger) { logger.Panic("baz") }, true, `{"level":"panic","msg":"baz"}`},
	} {
		withJSONLogger(t, DebugLevel, nil, func(logger Logger, buf *testBuffer) {
			if tc.should {
				assert.Panics(t, func() { tc.do(logger) }, "Expected panic")
			} else {
				assert.NotPanics(t, func() { tc.do(logger) }, "Expected no panic")
			}
			assert.Equal(t, tc.expected, buf.Stripped(), "Unexpected output from fatal-level Log.")
		})
	}
}

func TestJSONLoggerLogFatal(t *testing.T) {
	for _, tc := range []struct {
		do       func(Logger)
		should   bool
		status   int
		expected string
	}{
		{func(logger Logger) { logger.Check(FatalLevel, "bar").Write() }, true, 1, `{"level":"fatal","msg":"bar"}`},
		{func(logger Logger) { logger.Fatal("baz") }, true, 1, `{"level":"fatal","msg":"baz"}`},
	} {
		withJSONLogger(t, DebugLevel, nil, func(logger Logger, buf *testBuffer) {
			stub := stubExit()
			defer stub.Unstub()
			tc.do(logger)
			if tc.should {
				stub.AssertStatus(t, tc.status)
			} else {
				stub.AssertNoExit(t)
			}
			assert.Equal(t, tc.expected, buf.Stripped(), "Unexpected output from fatal-level Log.")
		})
	}
}

func TestJSONLoggerLeveledMethods(t *testing.T) {
	withJSONLogger(t, DebugLevel, nil, func(logger Logger, buf *testBuffer) {
		tests := []struct {
			method        func(string, ...Field)
			expectedLevel string
		}{
			{logger.Debug, "debug"},
			{logger.Info, "info"},
			{logger.Warn, "warn"},
			{logger.Error, "error"},
		}
		for _, tt := range tests {
			buf.Reset()
			tt.method("foo")
			expected := fmt.Sprintf(`{"level":"%s","msg":"foo"}`, tt.expectedLevel)
			assert.Equal(t, expected, buf.Stripped(), "Unexpected output from %s-level logger method.", tt.expectedLevel)
		}
	})
}

func TestJSONLoggerPanic(t *testing.T) {
	withJSONLogger(t, DebugLevel, nil, func(logger Logger, buf *testBuffer) {
		assert.Panics(t, func() { logger.Panic("foo") })
		assert.Equal(t, `{"level":"panic","msg":"foo"}`, buf.Stripped(), "Unexpected output from Logger.Panic.")
	})
}

func TestJSONLoggerCheckPanic(t *testing.T) {
	withJSONLogger(t, DebugLevel, nil, func(logger Logger, buf *testBuffer) {
		assert.Panics(t, func() {
			if cm := logger.Check(PanicLevel, "foo"); cm != nil {
				cm.Write()
			}
		})
		assert.Equal(t, `{"level":"panic","msg":"foo"}`, buf.Stripped(), "Unexpected output from Logger.Panic.")
	})
}

func TestJSONLoggerAlwaysPanics(t *testing.T) {
	withJSONLogger(t, FatalLevel, nil, func(logger Logger, buf *testBuffer) {
		assert.Panics(t, func() { logger.Panic("foo") }, "logger.Panics should panic")
		assert.Empty(t, buf.String(), "Panic should not be logged")
	})
}

func TestJSONLoggerCheckAlwaysPanics(t *testing.T) {
	withJSONLogger(t, FatalLevel, nil, func(logger Logger, buf *testBuffer) {
		assert.Panics(t, func() {
			if cm := logger.Check(PanicLevel, "foo"); cm != nil {
				cm.Write()
			}
		}, "CheckedMessage should panic")
		assert.Empty(t, buf.String(), "Panic should not be logged")
	})
}

func TestJSONLoggerFatal(t *testing.T) {
	stub := stubExit()
	defer stub.Unstub()

	withJSONLogger(t, DebugLevel, nil, func(logger Logger, buf *testBuffer) {
		logger.Fatal("foo")
		assert.Equal(t, `{"level":"fatal","msg":"foo"}`, buf.Stripped(), "Unexpected output from Logger.Fatal.")
		stub.AssertStatus(t, 1)
	})
}

func TestJSONLoggerCheckFatal(t *testing.T) {
	stub := stubExit()
	defer stub.Unstub()

	withJSONLogger(t, DebugLevel, nil, func(logger Logger, buf *testBuffer) {
		if cm := logger.Check(FatalLevel, "foo"); cm != nil {
			cm.Write()
		}
		assert.Equal(t, `{"level":"fatal","msg":"foo"}`, buf.Stripped(), "Unexpected output from Logger.Fatal.")
		stub.AssertStatus(t, 1)
	})
}

func TestJSONLoggerAlwaysFatals(t *testing.T) {
	stub := stubExit()
	defer stub.Unstub()

	withJSONLogger(t, FatalLevel+1, nil, func(logger Logger, buf *testBuffer) {
		logger.Fatal("foo")
		stub.AssertStatus(t, 1)
		assert.Empty(t, buf.String(), "Fatal should not be logged")
	})
}

func TestJSONLoggerCheckAlwaysFatals(t *testing.T) {
	stub := stubExit()
	defer stub.Unstub()

	withJSONLogger(t, FatalLevel+1, nil, func(logger Logger, buf *testBuffer) {
		if cm := logger.Check(FatalLevel, "foo"); cm != nil {
			cm.Write()
		}
		stub.AssertStatus(t, 1)
		assert.Empty(t, buf.String(), "Fatal should not be logged")
	})
}

func TestJSONLoggerDPanic(t *testing.T) {
	withJSONLogger(t, DebugLevel, nil, func(logger Logger, buf *testBuffer) {
		assert.NotPanics(t, func() { logger.DPanic("foo") })
		assert.Equal(t, `{"level":"dpanic","msg":"foo"}`, buf.Stripped(), "Unexpected output from DPanic in production mode.")
	})
	withJSONLogger(t, DebugLevel, opts(Development()), func(logger Logger, buf *testBuffer) {
		assert.Panics(t, func() { logger.DPanic("foo") })
		assert.Equal(t, `{"level":"dpanic","msg":"foo"}`, buf.Stripped(), "Unexpected output from Logger.Fatal in development mode.")
	})
}

func TestJSONLoggerNoOpsDisabledLevels(t *testing.T) {
	withJSONLogger(t, WarnLevel, nil, func(logger Logger, buf *testBuffer) {
		logger.Info("silence!")
		assert.Equal(t, []string{}, buf.Lines(), "Expected logging at a disabled level to produce no output.")
	})
}

func TestJSONLoggerWriteEntryFailure(t *testing.T) {
	errBuf := &testBuffer{}
	errSink := &spywrite.WriteSyncer{Writer: errBuf}
	logger := New(
		WriterFacility(newJSONEncoder(), AddSync(spywrite.FailWriter{}), DebugLevel),
		ErrorOutput(errSink))

	logger.Info("foo")
	// Should log the error.
	assert.Regexp(t, `write error: failed`, errBuf.Stripped(), "Expected to log the error to the error output.")
	assert.True(t, errSink.Called(), "Expected logging an internal error to call Sync the error sink.")
}

func TestJSONLoggerSyncsOutput(t *testing.T) {
	sink := &spywrite.WriteSyncer{Writer: ioutil.Discard}
	logger := New(WriterFacility(newJSONEncoder(), sink, DebugLevel))

	logger.Error("foo")
	assert.False(t, sink.Called(), "Didn't expect logging at error level to Sync underlying WriteCloser.")

	assert.Panics(t, func() { logger.Panic("foo") }, "Expected panic when logging at Panic level.")
	assert.True(t, sink.Called(), "Expected logging at panic level to Sync underlying WriteSyncer.")
}

func TestLoggerAddCaller(t *testing.T) {
	withJSONLogger(t, DebugLevel, opts(AddCaller()), func(logger Logger, buf *testBuffer) {
		logger.Info("Callers.")
		assert.Regexp(t,
			`"caller":"[^"]+/logger_test.go:\d+","msg":"Callers\."`,
			buf.Stripped(), "Expected to find package name and file name in output.")
	})
}

func TestLoggerAddCallerSkip(t *testing.T) {
	withJSONLogger(t, DebugLevel, opts(AddCaller(), AddCallerSkip(1)), func(logger Logger, buf *testBuffer) {
		logger.Info("Callers.")
		assert.Regexp(t,
			`"caller":"[^"]+/common_test.go:\d+","msg":"Callers\."`,
			buf.Stripped(), "Expected to find package name and file name in output.")
	})
}

func TestLoggerAddCallerFail(t *testing.T) {
	errBuf := &testBuffer{}
	withJSONLogger(t, DebugLevel, opts(
		AddCaller(),
		ErrorOutput(errBuf),
	), func(log Logger, buf *testBuffer) {
		logImpl := log.(*logger)
		logImpl.callerSkip = 1e3

		log.Info("Failure.")
		assert.Regexp(t,
			`addCaller error: failed to get caller`,
			errBuf.String(), "Didn't find expected failure message.")
		assert.Contains(t,
			buf.String(), `"msg":"Failure."`,
			"Expected original message to survive failures in runtime.Caller.")
	})
}

func TestLoggerAddStacks(t *testing.T) {
	withJSONLogger(t, DebugLevel, opts(AddStacks(InfoLevel)), func(logger Logger, buf *testBuffer) {
		logger.Info("Stacks.")
		output := buf.String()
		require.Contains(t, output, "zap.TestLoggerAddStacks", "Expected to find test function in stacktrace.")
		assert.Contains(t, output, `"stacktrace":`, "Stacktrace added under an unexpected key.")

		buf.Reset()
		logger.Warn("Stacks.")
		assert.Contains(t, buf.String(), `"stacktrace":`, "Expected to include stacktrace at Warn level.")

		buf.Reset()
		logger.Debug("No stacks.")
		assert.NotContains(t, buf.String(), "Unexpected stacktrace at Debug level.")
	})
}

func TestLoggerConcurrent(t *testing.T) {
	withJSONLogger(t, DebugLevel, nil, func(logger Logger, buf *testBuffer) {
		child := logger.With(String("foo", "bar"))

		wg := &sync.WaitGroup{}
		runConcurrently(5 /* goroutines */, 10 /* iterations */, wg, func() {
			logger.Info("info", String("foo", "bar"))
		})
		runConcurrently(5 /* goroutines */, 10 /* iterations */, wg, func() {
			child.Info("info")
		})

		wg.Wait()

		// Make sure the output doesn't contain interspersed entries.
		expected := `{"level":"info","msg":"info","foo":"bar"}` + "\n"
		assert.Equal(t, strings.Repeat(expected, 100), buf.String())
	})
}
