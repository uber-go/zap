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

package zapcore

import (
	"bytes"
	"errors"
	"io/ioutil"
	"strings"
)

// A TestSyncer is a spy for the Sync portion of WriteTestSyncer.
type TestSyncer struct {
	err    error
	called bool
}

// SetError sets the error that the Sync method will return.
func (s *TestSyncer) SetError(err error) {
	s.err = err
}

// Sync records that it was called, then returns the user-supplied error (if
// any).
func (s *TestSyncer) Sync() error {
	s.called = true
	return s.err
}

// Called reports whether the Sync method was called.
func (s *TestSyncer) Called() bool {
	return s.called
}

// A TestDiscarder sends all writes to ioutil.Discard.
type TestDiscarder struct{ TestSyncer }

// Push implements Pusher.
func (d *TestDiscarder) Push(_ Level, b []byte) (int, error) {
	return ioutil.Discard.Write(b)
}

// TestFailPusher is a Pusher that always returns an error on writes.
type TestFailPusher struct{ TestSyncer }

// Push implements Pusher.
func (w *TestFailPusher) Push(_ Level, b []byte) (int, error) {
	return len(b), errors.New("failed")
}

// TestShortPusher is a Pusher whose write method never fails, but
// nevertheless fails to the last byte of the input.
type TestShortPusher struct{ TestSyncer }

// Push implements Pusher.
func (w *TestShortPusher) Push(_ Level, b []byte) (int, error) {
	return len(b) - 1, nil
}

// TestBuffer is an implementation of Pusher that sends all writes to
// a bytes.TestBuffer. It has convenience methods to split the accumulated buffer
// on newlines.
type TestBuffer struct {
	bytes.Buffer
	TestSyncer
}

// Push implements Pusher.
func (b *TestBuffer) Push(_ Level, p []byte) (int, error) {
	return b.Write(p)
}

// Lines returns the current buffer contents, split on newlines.
func (b *TestBuffer) Lines() []string {
	output := strings.Split(b.String(), "\n")
	return output[:len(output)-1]
}

// Stripped returns the current buffer contents with the last trailing newline
// stripped.
func (b *TestBuffer) Stripped() string {
	return strings.TrimRight(b.String(), "\n")
}
