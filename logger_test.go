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

func TestJSONLoggerSetLevel(t *testing.T) {
	withJSONLogger(t, nil, func(logger Logger, _ *testBuffer) {
		assert.Equal(t, DebugLevel, logger.Level(), "Unexpected initial level.")
		logger.SetLevel(ErrorLevel)
		assert.Equal(t, ErrorLevel, logger.Level(), "Unexpected level after SetLevel.")
	})
}

func TestJSONLoggerRuntimeLevelChange(t *testing.T) {
	// Test that changing a logger's level also changes the level of all
	// ancestors and descendants.
	grandparent := New(newJSONEncoder(), Fields(Int("generation", 1)))
	parent := grandparent.With(Int("generation", 2))
	child := parent.With(Int("generation", 3))

	all := []Logger{grandparent, parent, child}
	for _, logger := range all {
		assert.Equal(t, InfoLevel, logger.Level(), "Expected all loggers to start at Info level.")
	}

	parent.SetLevel(DebugLevel)
	for _, logger := range all {
		assert.Equal(t, DebugLevel, logger.Level(), "Expected all loggers to switch to Debug level.")
	}
}

func TestJSONLoggerConcurrentLevelMutation(t *testing.T) {
	// Trigger races for non-atomic level mutations.
	logger := New(newJSONEncoder())

	proceed := make(chan struct{})
	wg := &sync.WaitGroup{}
	runConcurrently(10, 100, wg, func() {
		<-proceed
		logger.Level()
	})
	runConcurrently(10, 100, wg, func() {
		<-proceed
		logger.SetLevel(WarnLevel)
	})
	close(proceed)
	wg.Wait()
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

	withJSONLogger(t, nil, func(logger Logger, buf *testBuffer) {
		assert.Panics(t, func() { logger.Log(PanicLevel, "foo") }, "Expected logging at Panic level to panic.")
		assert.Equal(t, `{"level":"panic","msg":"foo"}`, buf.Stripped(), "Unexpected output from panic-level Log.")
	})

	stub := stubExit()
	defer stub.Unstub()
	withJSONLogger(t, nil, func(logger Logger, buf *testBuffer) {
		logger.Log(FatalLevel, "foo")
		assert.Equal(t, `{"level":"fatal","msg":"foo"}`, buf.Stripped(), "Unexpected output from fatal-level Log.")
		stub.AssertStatus(t, 1)
	})
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
	withJSONLogger(t, nil, func(logger Logger, buf *testBuffer) {
		logger.SetLevel(WarnLevel)
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
	assert.Equal(t, "failed", errBuf.Stripped(), "Expected to log the error to the error output.")
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
