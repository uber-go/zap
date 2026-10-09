// Copyright (c) 2017 Uber Technologies, Inc.
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

package zaptest

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/internal/ztest"
	"go.uber.org/zap/zapcore"

	"github.com/stretchr/testify/assert"
)

func TestTestLogger(t *testing.T) {
	ts := newTestLogSpy(t)
	defer ts.AssertPassed()

	log := NewLogger(ts)

	log.Info("received work order")
	log.Debug("starting work")
	log.Warn("work may fail")
	log.Error("work failed", zap.Error(errors.New("great sadness")))

	assert.Panics(t, func() {
		log.Panic("failed to do work")
	}, "log.Panic should panic")

	ts.AssertMessages(
		"INFO	received work order",
		"DEBUG	starting work",
		"WARN	work may fail",
		`ERROR	work failed	{"error": "great sadness"}`,
		"PANIC	failed to do work",
	)
}

func TestTestLoggerSupportsLevels(t *testing.T) {
	ts := newTestLogSpy(t)
	defer ts.AssertPassed()

	log := NewLogger(ts, Level(zap.WarnLevel))

	log.Info("received work order")
	log.Debug("starting work")
	log.Warn("work may fail")
	log.Error("work failed", zap.Error(errors.New("great sadness")))

	assert.Panics(t, func() {
		log.Panic("failed to do work")
	}, "log.Panic should panic")

	ts.AssertMessages(
		"WARN	work may fail",
		`ERROR	work failed	{"error": "great sadness"}`,
		"PANIC	failed to do work",
	)
}

func TestTestLoggerSupportsWrappedZapOptions(t *testing.T) {
	ts := newTestLogSpy(t)
	defer ts.AssertPassed()

	log := NewLogger(ts, WrapOptions(zap.AddCaller(), zap.Fields(zap.String("k1", "v1"))))

	log.Info("received work order")
	log.Debug("starting work")
	log.Warn("work may fail")
	log.Error("work failed", zap.Error(errors.New("great sadness")))

	assert.Panics(t, func() {
		log.Panic("failed to do work")
	}, "log.Panic should panic")

	ts.AssertMessages(
		`INFO	zaptest/logger_test.go:89	received work order	{"k1": "v1"}`,
		`DEBUG	zaptest/logger_test.go:90	starting work	{"k1": "v1"}`,
		`WARN	zaptest/logger_test.go:91	work may fail	{"k1": "v1"}`,
		`ERROR	zaptest/logger_test.go:92	work failed	{"k1": "v1", "error": "great sadness"}`,
		`PANIC	zaptest/logger_test.go:95	failed to do work	{"k1": "v1"}`,
	)
}

func TestTestingWriter(t *testing.T) {
	ts := newTestLogSpy(t)
	w := NewTestingWriter(ts)

	n, err := io.WriteString(w, "hello\n\n")
	assert.NoError(t, err, "WriteString must not fail")
	assert.Equal(t, 7, n)
}

func TestTestLoggerErrorOutput(t *testing.T) {
	// This test verifies that the test logger logs internal messages to the
	// testing.T and marks the test as failed.

	ts := newTestLogSpy(t)
	defer ts.AssertFailed()

	log := NewLogger(ts)

	// Replace with a core that fails.
	log = log.WithOptions(zap.WrapCore(func(zapcore.Core) zapcore.Core {
		return zapcore.NewCore(
			zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()),
			zapcore.Lock(zapcore.AddSync(ztest.FailWriter{})),
			zapcore.DebugLevel,
		)
	}))

	log.Info("foo") // this fails

	if assert.Len(t, ts.Messages, 1, "expected a log message") {
		assert.Regexp(t, `write error: failed`, ts.Messages[0])
	}
}

// testLogSpy is a testing.TB that captures logged messages.
type testLogSpy struct {
	testing.TB

	failed   bool
	Messages []string
}

func newTestLogSpy(t testing.TB) *testLogSpy {
	return &testLogSpy{TB: t}
}

func (t *testLogSpy) Fail() {
	t.failed = true
}

func (t *testLogSpy) Failed() bool {
	return t.failed
}

func (t *testLogSpy) FailNow() {
	t.Fail()
	t.TB.FailNow()
}

func (t *testLogSpy) Logf(format string, args ...interface{}) {
	// Log messages are in the format,
	//
	//   2017-10-27T13:03:01.000-0700	DEBUG	your message here	{data here}
	//
	// We strip the first part of these messages because we can't really test
	// for the timestamp from these tests.
	m := fmt.Sprintf(format, args...)
	m = m[strings.IndexByte(m, '\t')+1:]
	t.Messages = append(t.Messages, m)
	t.Log(m)
}

func (t *testLogSpy) AssertMessages(msgs ...string) {
	assert.Equal(t.TB, msgs, t.Messages, "logged messages did not match")
}

func (t *testLogSpy) AssertPassed() {
	t.assertFailed(false, "expected test to pass")
}

func (t *testLogSpy) AssertFailed() {
	t.assertFailed(true, "expected test to fail")
}

func (t *testLogSpy) assertFailed(v bool, msg string) {
	assert.Equal(t.TB, v, t.failed, msg)
}

func TestTestLoggerAfterTestCompletedPanics(t *testing.T) {
	w := newTestFinishedWrapper(t)
	log := NewLogger(w)

	assert.Panics(t, func() {
		log.Info("foo")
	})
}

func TestTestLoggerWithMuteAfterTestCompletion(t *testing.T) {
	w := newTestFinishedWrapper(t)
	log := NewLogger(w, MuteAfterTestCompletion())

	assert.NotPanics(t, func() {
		log.Info("first call after test completed")
		log.Info("second call after test completed")
	})

	assert.Equal(t, 1, w.LogCount(), "subsequent logs should be muted early without calling Logf again")
}

func TestTestLoggerWithMuteAfterTestCompletion_Concurrent(t *testing.T) {
	w := newTestFinishedWrapper(t)
	log := NewLogger(w, MuteAfterTestCompletion())

	const goroutines = 20
	done := make(chan struct{}, goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer func() { done <- struct{}{} }()
			assert.NotPanics(t, func() {
				log.Info("concurrent log after test completed")
			})
		}()
	}

	for i := 0; i < goroutines; i++ {
		<-done
	}
}

func TestTestLoggerWithMuteAfterTestCompletion_OtherPanic(t *testing.T) {
	w := newTestFinishedWrapper(t)
	w.logPanicMsg = "unexpected runtime panic"
	log := NewLogger(w, MuteAfterTestCompletion())

	assert.PanicsWithValue(t, "unexpected runtime panic", func() {
		log.Info("foo")
	})
}

func TestTestingWriterWithMuteAfterTestCompletion(t *testing.T) {
	w := newTestFinishedWrapper(t)

	unmuted := NewTestingWriter(w)
	assert.Panics(t, func() {
		_, _ = unmuted.Write([]byte("foo\n"))
	})

	muted := NewTestingWriter(w).WithMuteAfterTestCompletion(true)
	assert.NotPanics(t, func() {
		n, err := muted.Write([]byte("foo\n"))
		assert.NoError(t, err)
		assert.Equal(t, 4, n)
	})
}

func TestTestLoggerErrorOutputAfterTestCompleted(t *testing.T) {
	w := newTestFinishedWrapper(t)
	log := NewLogger(w, MuteAfterTestCompletion())

	log = log.WithOptions(zap.WrapCore(func(zapcore.Core) zapcore.Core {
		return zapcore.NewCore(
			zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()),
			zapcore.Lock(zapcore.AddSync(ztest.FailWriter{})),
			zapcore.DebugLevel,
		)
	}))

	assert.NotPanics(t, func() {
		log.Info("foo")
	})
}

func TestTestingWriterFailAfterTestCompletedPanics(t *testing.T) {
	w := newTestFinishedWrapper(t)
	w.logPanicMsg = ""
	w.failPanicMsg = "Fail in goroutine after " + t.Name() + " has completed"

	unmuted := NewTestingWriter(w).WithMarkFailed(true)
	assert.Panics(t, func() {
		_, _ = unmuted.Write([]byte("foo\n"))
	})

	muted := NewTestingWriter(w).WithMarkFailed(true).WithMuteAfterTestCompletion(true)
	assert.NotPanics(t, func() {
		n, err := muted.Write([]byte("foo\n"))
		assert.NoError(t, err)
		assert.Equal(t, 4, n)
	})
}

func TestTestLoggerNormalLoggingWithMute(t *testing.T) {
	ts := newTestLogSpy(t)
	defer ts.AssertPassed()

	log := NewLogger(ts, MuteAfterTestCompletion())
	log.Info("regular message")

	ts.AssertMessages("INFO\tregular message")
}

type testFinishedWrapper struct {
	TestingT
	logPanicMsg  string
	failPanicMsg string
	logged       chan struct{}
}

func newTestFinishedWrapper(t TestingT) *testFinishedWrapper {
	return &testFinishedWrapper{
		TestingT:    t,
		logPanicMsg: "Log in goroutine after " + t.Name() + " has completed",
		logged:      make(chan struct{}, 100),
	}
}

func (f *testFinishedWrapper) LogCount() int {
	return len(f.logged)
}

func (f *testFinishedWrapper) Logf(format string, args ...interface{}) {
	f.logged <- struct{}{}
	if f.logPanicMsg != "" {
		panic(f.logPanicMsg)
	}
	f.TestingT.Logf(format, args...)
}

func (f *testFinishedWrapper) Fail() {
	if f.failPanicMsg != "" {
		panic(f.failPanicMsg)
	}
	f.TestingT.Fail()
}
