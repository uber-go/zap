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
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/uber-go/zap/spywrite"

	"github.com/stretchr/testify/assert"
)

func opts(opts ...Option) []Option {
	return opts
}

type stubbedExit struct {
	Status *int
}

func (se *stubbedExit) Unstub() {
	_exit = os.Exit
}

func (se *stubbedExit) AssertNoExit(t testing.TB) {
	assert.Nil(t, se.Status, "Unexpected exit.")
}

func (se *stubbedExit) AssertStatus(t testing.TB, expected int) {
	if assert.NotNil(t, se.Status, "Expected to exit.") {
		assert.Equal(t, expected, *se.Status, "Unexpected exit code.")
	}
}

func stubExit() *stubbedExit {
	stub := &stubbedExit{}
	_exit = func(s int) { stub.Status = &s }
	return stub
}

func withJSONLogger(t testing.TB, opts []Option, f func(Logger, *testBuffer)) {
	sink := &testBuffer{}
	errSink := &testBuffer{}

	allOpts := make([]Option, 0, 3+len(opts))
	allOpts = append(allOpts, DebugLevel, Output(sink), ErrorOutput(errSink))
	allOpts = append(allOpts, opts...)
	logger := New(newJSONEncoder(NoTime()), allOpts...)

	f(logger, sink)
	assert.Empty(t, errSink.String(), "Expected error sink to be empty.")
}

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
	withJSONLogger(t, opts(dl), func(grandparent Logger, buf *testBuffer) {
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
	withJSONLogger(t, fieldOpts, func(logger Logger, buf *testBuffer) {
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
	withJSONLogger(t, fieldOpts, func(logger Logger, buf *testBuffer) {
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

func TestJSONLoggerLog(t *testing.T) {
	withJSONLogger(t, nil, func(logger Logger, buf *testBuffer) {
		logger.Log(DebugLevel, "foo")
		assert.Equal(t, `{"level":"debug","msg":"foo"}`, buf.Stripped(), "Unexpected output from Log.")
	})
}

func TestJSONLoggerLogPanic(t *testing.T) {
	for _, tc := range []struct {
		do       func(Logger)
		should   bool
		expected string
	}{
		{func(logger Logger) { logger.Log(PanicLevel, "foo") }, false, `{"level":"panic","msg":"foo"}`},
		{func(logger Logger) { logger.Check(PanicLevel, "bar").Write() }, true, `{"level":"panic","msg":"bar"}`},
		{func(logger Logger) { logger.Panic("baz") }, true, `{"level":"panic","msg":"baz"}`},
	} {
		withJSONLogger(t, nil, func(logger Logger, buf *testBuffer) {
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
		{func(logger Logger) { logger.Log(FatalLevel, "foo") }, false, 0, `{"level":"fatal","msg":"foo"}`},
		{func(logger Logger) { logger.Check(FatalLevel, "bar").Write() }, true, 1, `{"level":"fatal","msg":"bar"}`},
		{func(logger Logger) { logger.Fatal("baz") }, true, 1, `{"level":"fatal","msg":"baz"}`},
	} {
		withJSONLogger(t, nil, func(logger Logger, buf *testBuffer) {
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
	withJSONLogger(t, nil, func(logger Logger, buf *testBuffer) {
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
	withJSONLogger(t, nil, func(logger Logger, buf *testBuffer) {
		assert.Panics(t, func() { logger.Panic("foo") })
		assert.Equal(t, `{"level":"panic","msg":"foo"}`, buf.Stripped(), "Unexpected output from Logger.Panic.")
	})
}

func TestJSONLoggerCheckPanic(t *testing.T) {
	withJSONLogger(t, nil, func(logger Logger, buf *testBuffer) {
		assert.Panics(t, func() {
			if cm := logger.Check(PanicLevel, "foo"); cm.OK() {
				cm.Write()
			}
		})
		assert.Equal(t, `{"level":"panic","msg":"foo"}`, buf.Stripped(), "Unexpected output from Logger.Panic.")
	})
}

func TestJSONLoggerAlwaysPanics(t *testing.T) {
	withJSONLogger(t, []Option{FatalLevel}, func(logger Logger, buf *testBuffer) {
		assert.Panics(t, func() { logger.Panic("foo") }, "logger.Panics should panic")
		assert.Empty(t, buf.String(), "Panic should not be logged")
	})
}

func TestJSONLoggerCheckAlwaysPanics(t *testing.T) {
	withJSONLogger(t, []Option{FatalLevel}, func(logger Logger, buf *testBuffer) {
		assert.Panics(t, func() {
			if cm := logger.Check(PanicLevel, "foo"); cm.OK() {
				cm.Write()
			}
		}, "CheckedMessage should panic")
		assert.Empty(t, buf.String(), "Panic should not be logged")
	})
}

func TestJSONLoggerFatal(t *testing.T) {
	stub := stubExit()
	defer stub.Unstub()

	withJSONLogger(t, nil, func(logger Logger, buf *testBuffer) {
		logger.Fatal("foo")
		assert.Equal(t, `{"level":"fatal","msg":"foo"}`, buf.Stripped(), "Unexpected output from Logger.Fatal.")
		stub.AssertStatus(t, 1)
	})
}

func TestJSONLoggerCheckFatal(t *testing.T) {
	stub := stubExit()
	defer stub.Unstub()

	withJSONLogger(t, nil, func(logger Logger, buf *testBuffer) {
		if cm := logger.Check(FatalLevel, "foo"); cm.OK() {
			cm.Write()
		}
		assert.Equal(t, `{"level":"fatal","msg":"foo"}`, buf.Stripped(), "Unexpected output from Logger.Fatal.")
		stub.AssertStatus(t, 1)
	})
}

func TestJSONLoggerAlwaysFatals(t *testing.T) {
	stub := stubExit()
	defer stub.Unstub()

	withJSONLogger(t, []Option{FatalLevel + 1}, func(logger Logger, buf *testBuffer) {
		logger.Fatal("foo")
		stub.AssertStatus(t, 1)
		assert.Empty(t, buf.String(), "Fatal should not be logged")
	})
}

func TestJSONLoggerCheckAlwaysFatals(t *testing.T) {
	stub := stubExit()
	defer stub.Unstub()

	withJSONLogger(t, []Option{FatalLevel + 1}, func(logger Logger, buf *testBuffer) {
		if cm := logger.Check(FatalLevel, "foo"); cm.OK() {
			cm.Write()
		}
		stub.AssertStatus(t, 1)
		assert.Empty(t, buf.String(), "Fatal should not be logged")
	})
}

func TestJSONLoggerDFatal(t *testing.T) {
	stub := stubExit()
	defer stub.Unstub()

	withJSONLogger(t, nil, func(logger Logger, buf *testBuffer) {
		logger.DFatal("foo")
		assert.Equal(t, `{"level":"error","msg":"foo"}`, buf.Stripped(), "Unexpected output from DFatal in production mode.")
		stub.AssertNoExit(t)
	})
	withJSONLogger(t, opts(Development()), func(logger Logger, buf *testBuffer) {
		logger.DFatal("foo")
		assert.Equal(t, `{"level":"fatal","msg":"foo"}`, buf.Stripped(), "Unexpected output from Logger.Fatal in development mode.")
		stub.AssertStatus(t, 1)
	})
}

func TestJSONLoggerNoOpsDisabledLevels(t *testing.T) {
	withJSONLogger(t, opts(WarnLevel), func(logger Logger, buf *testBuffer) {
		logger.Info("silence!")
		assert.Equal(t, []string{}, buf.Lines(), "Expected logging at a disabled level to produce no output.")
	})
}

func TestJSONLoggerWriteEntryFailure(t *testing.T) {
	errBuf := &testBuffer{}
	errSink := &spywrite.WriteSyncer{Writer: errBuf}
	logger := New(
		newJSONEncoder(),
		DebugLevel,
		Output(AddSync(spywrite.FailWriter{})),
		ErrorOutput(errSink),
	)

	logger.Info("foo")
	// Should log the error.
	assert.Regexp(t, `encoder error: failed`, errBuf.Stripped(), "Expected to log the error to the error output.")
	assert.True(t, errSink.Called(), "Expected logging an internal error to call Sync the error sink.")
}

func TestJSONLoggerSyncsOutput(t *testing.T) {
	sink := &spywrite.WriteSyncer{Writer: ioutil.Discard}
	logger := New(newJSONEncoder(), DebugLevel, Output(sink))

	logger.Error("foo")
	assert.False(t, sink.Called(), "Didn't expect logging at error level to Sync underlying WriteCloser.")

	assert.Panics(t, func() { logger.Panic("foo") }, "Expected panic when logging at Panic level.")
	assert.True(t, sink.Called(), "Expected logging at panic level to Sync underlying WriteSyncer.")
}

func TestLoggerConcurrent(t *testing.T) {
	withJSONLogger(t, nil, func(logger Logger, buf *testBuffer) {
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

func runConcurrently(goroutines, iterations int, wg *sync.WaitGroup, f func()) {
	wg.Add(goroutines)
	for g := 0; g < goroutines; g++ {
		go func() {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				f()
			}
		}()
	}
}
