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

package spy

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

// WriteSyncer is a concrete type that implements zap.WriteSyncer.
type WriteSyncer struct {
	io.Writer
	Err        error
	SyncCalled bool
}

// Sync sets the SyncCalled bit and returns the user-specified error.
func (w *WriteSyncer) Sync() error {
	w.SyncCalled = true
	return w.Err
}

// WriteFlusher is a concrete type that implements zap.WriteFlusher.
type WriteFlusher struct {
	io.Writer
	Err         error
	FlushCalled bool
}

// Flush sets the FlushCalled bit and returns the user-specified error.
func (w *WriteFlusher) Flush() error {
	w.FlushCalled = true
	return w.Err
}
