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
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func stubTime() func() {
	_timeNow = func() time.Time { return time.Unix(0, 0) }
	return func() { _timeNow = time.Now }
}

func withJSONLogger(f func(*jsonLogger, func() []string), fields ...Field) {
	defer stubTime()()
	sink := bytes.NewBuffer(nil)
	jl := NewJSON(All, sink, fields...)

	f(jl.(*jsonLogger), func() []string { return strings.Split(sink.String(), "\n") })
}

func assertMessage(t testing.TB, level, expectedMsg, actualLog string) {
	expectedLog := fmt.Sprintf(`{"msg":"%s","level":"%s","ts":0,"fields":{}}`, expectedMsg, level)
	assert.Equal(t, expectedLog, actualLog, "Unexpected log output.")
}

func assertFields(t testing.TB, jl Logger, getOutput func() []string, expectedFields ...string) {
	jl.Debug("")
	actualLogs := getOutput()
	for i, fields := range expectedFields {
		expectedLog := fmt.Sprintf(`{"msg":"","level":"debug","ts":0,"fields":%s}`, fields)
		assert.Equal(t, expectedLog, actualLogs[i], "Unexpected log output.")
	}
}

func TestJSONLoggerSetLevel(t *testing.T) {
	withJSONLogger(func(jl *jsonLogger, _ func() []string) {
		assert.Equal(t, All, jl.Level(), "Unexpected initial level.")
		jl.SetLevel(Debug)
		assert.Equal(t, Debug, jl.Level(), "Unexpected level after SetLevel.")
	})
}

func TestJSONLoggerEnabled(t *testing.T) {
	withJSONLogger(func(jl *jsonLogger, _ func() []string) {
		jl.SetLevel(Info)
		assert.False(t, jl.Enabled(Debug), "Debug logs shouldn't be enabled at Info level.")
		assert.True(t, jl.Enabled(Info), "Info logs should be enabled at Info level.")
		assert.True(t, jl.Enabled(Warn), "Warn logs should be enabled at Info level.")
		assert.True(t, jl.Enabled(Error), "Error logs should be enabled at Info level.")
		assert.True(t, jl.Enabled(Panic), "Panic logs should be enabled at Info level.")
		assert.True(t, jl.Enabled(Fatal), "Fatal logs should be enabled at Info level.")

		for _, lvl := range []Level{Debug, Info, Warn, Error, Panic, Fatal} {
			jl.SetLevel(None)
			assert.False(t, jl.Enabled(lvl), "No logging should be enabled at None level.")
			jl.SetLevel(All)
			assert.True(t, jl.Enabled(lvl), "All logging should be enabled at All level.")
		}
	})
}

func TestJSONLoggerConcurrentLevelMutation(t *testing.T) {
	// Trigger races for non-atomic level mutations.
	proceed := make(chan struct{})
	jl := NewJSON(Info, ioutil.Discard)

	for i := 0; i < 50; i++ {
		go func(l Logger) {
			<-proceed
			jl.Level()
		}(jl)
		go func(l Logger) {
			<-proceed
			jl.SetLevel(Warn)
		}(jl)
	}
	close(proceed)
}

func TestJSONLoggerInitialFields(t *testing.T) {
	withJSONLogger(func(jl *jsonLogger, output func() []string) {
		assertFields(t, jl, output, `{"foo":42,"bar":"baz"}`)
	}, Int("foo", 42), String("bar", "baz"))
}

func TestJSONLoggerWith(t *testing.T) {
	withJSONLogger(func(jl *jsonLogger, output func() []string) {
		// Child loggers should have copy-on-write semantics, so two children
		// shouldn't stomp on each other's fields or affect the parent's fields.
		jl.With(String("one", "two")).Debug("")
		jl.With(String("three", "four")).Debug("")
		assertFields(t, jl, output, `{"foo":42,"one":"two"}`, `{"foo":42,"three":"four"}`, `{"foo":42}`)
	}, Int("foo", 42))
}

func TestJSONLoggerDebug(t *testing.T) {
	withJSONLogger(func(jl *jsonLogger, output func() []string) {
		jl.Debug("foo")
		assertMessage(t, "debug", "foo", output()[0])
	})
}

func TestJSONLoggerInfo(t *testing.T) {
	withJSONLogger(func(jl *jsonLogger, output func() []string) {
		jl.Info("foo")
		assertMessage(t, "info", "foo", output()[0])
	})
}

func TestJSONLoggerWarn(t *testing.T) {
	withJSONLogger(func(jl *jsonLogger, output func() []string) {
		jl.Warn("foo")
		assertMessage(t, "warn", "foo", output()[0])
	})
}

func TestJSONLoggerError(t *testing.T) {
	withJSONLogger(func(jl *jsonLogger, output func() []string) {
		jl.Error("foo")
		assertMessage(t, "error", "foo", output()[0])
	})
}

func TestJSONLoggerPanic(t *testing.T) {
	withJSONLogger(func(jl *jsonLogger, output func() []string) {
		assert.Panics(t, func() {
			jl.Panic("foo")
		})
		assertMessage(t, "panic", "foo", output()[0])
	})
}

func TestJSONLoggerNoOpsDisabledLevels(t *testing.T) {
	withJSONLogger(func(jl *jsonLogger, output func() []string) {
		jl.SetLevel(Warn)
		jl.Info("silence!")
		assert.Equal(t, []string{""}, output(), "Expected logging at a disabled level to produce no output.")
	})
}

func TestJSONLoggerInternalErrorHandling(t *testing.T) {
	defer stubTime()()

	errBuf := bytes.NewBuffer(nil)
	_errSink = errBuf
	defer func() { _errSink = os.Stderr }()

	buf := bytes.NewBuffer(nil)

	jl := NewJSON(All, buf, Object("user", fakeUser{"fail"}))
	output := func() []string { return strings.Split(buf.String(), "\n") }

	// Expect partial output, even if there's an error serializing
	// user-defined types.
	assertFields(t, jl, output, `{"user":{}}`)
	// Internal errors go to stderr.
	assert.Equal(t, "fail", errBuf.String(), "Expected internal errors to print to stderr.")
}
