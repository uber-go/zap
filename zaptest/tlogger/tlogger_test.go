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

package tlogger

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"

	"go.uber.org/zap"

	"github.com/stretchr/testify/assert"
)

func TestTestLoggerIncludesDebug(t *testing.T) {
	ts := newTestLogSpy(t)
	log := zap.New(New(ts))
	log.Debug("calculating")
	log.Info("finished calculating", zap.Int("answer", 42))

	ts.AssertMessages(
		"DEBUG	calculating",
		`INFO	finished calculating	{"answer": 42}`,
	)
}

func TestTestLoggerAt(t *testing.T) {
	ts := newTestLogSpy(t)
	log := zap.New(NewAt(ts, zap.WarnLevel))

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

// testLogSpy is a testing.TB that captures logged messages.
type testLogSpy struct {
	testing.TB

	mu       sync.Mutex
	Messages []string
}

func newTestLogSpy(t testing.TB) *testLogSpy {
	return &testLogSpy{TB: t}
}

func (t *testLogSpy) Log(args ...interface{}) {
	// Log messages are in the format,
	//
	//   2017-10-27T13:03:01.000-0700	DEBUG	your message here	{data here}
	//
	// We strip the first part of these messages because we can't really test
	// for the timestamp from these tests.
	m := fmt.Sprint(args...)
	m = m[strings.IndexByte(m, '\t')+1:]

	// t.Log should be thread-safe.
	t.mu.Lock()
	t.Messages = append(t.Messages, m)
	t.mu.Unlock()

	t.TB.Log(args...)
}

func (t *testLogSpy) AssertMessages(msgs ...string) {
	assert.Equal(t, msgs, t.Messages)
}
