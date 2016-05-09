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

package spywrite

import (
	"errors"
	"io"
)

// FailWriter is an io.Writer that always returns an error.
type FailWriter struct{}

// Write implements io.Writer.
func (w FailWriter) Write(b []byte) (int, error) {
	return len(b), errors.New("failed")
}

// ShortWriter is an io.Writer that never returns an error, but doesn't write
// the last byte of the input.
type ShortWriter struct{}

// Write implements io.Writer.
func (w ShortWriter) Write(b []byte) (int, error) {
	return len(b) - 1, nil
}

// A Syncer is a spy for the Sync portion of zap.WriteSyncer.
type Syncer struct {
	err    error
	called bool
}

// SetError sets the error that the Sync method will return.
func (s *Syncer) SetError(err error) {
	s.err = err
}

// Sync records that it was called, then returns the user-supplied error (if
// any).
func (s *Syncer) Sync() error {
	s.called = true
	return s.err
}

// Called reports whether the Sync method was called.
func (s *Syncer) Called() bool {
	return s.called
}

// A Flusher is a spy for the Flush portion of zap.WriteFlusher.
type Flusher struct {
	err    error
	called bool
}

// SetError sets the error that the Flush method will return.
func (f *Flusher) SetError(err error) {
	f.err = err
}

// Flush records that it was called, then returns the user-supplied error (if
// any).
func (f *Flusher) Flush() error {
	f.called = true
	return f.err
}

// Called reports whether the Flush method was called.
func (f *Flusher) Called() bool {
	return f.called
}

// WriteSyncer is a concrete type that implements zap.WriteSyncer.
type WriteSyncer struct {
	io.Writer
	Syncer
}

// WriteFlusher is a concrete type that implements zap.WriteFlusher.
type WriteFlusher struct {
	io.Writer
	Flusher
}

// A WriteFlushSyncer implements both zap.WriteFlusher and zap.WriteSyncer.
type WriteFlushSyncer struct {
	io.Writer
	Syncer
	Flusher
}
