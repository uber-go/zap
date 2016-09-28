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
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTeeWritesBoth(t *testing.T) {
	first := &bytes.Buffer{}
	second := &bytes.Buffer{}
	ws := Tee(AddSync(first), AddSync(second))

	msg := []byte("dumbledore")
	n, err := ws.Write(msg)
	require.NoError(t, err, "Expected successful buffer write")
	assert.Equal(t, len(msg), n)

	assert.Equal(t, msg, first.Bytes())
	assert.Equal(t, msg, second.Bytes())
}

func TestTeeFailsWrite(t *testing.T) {
	failer := failingWriter{err: errors.New("can't touch this")}
	ws := Tee(AddSync(failer))

	_, err := ws.Write([]byte("test"))
	assert.Error(t, err, "Write error should propagate")
}

func TestTeeFailsShortWrite(t *testing.T) {
	shorter := failingWriter{nBytes: 1}
	ws := Tee(AddSync(shorter))

	n, err := ws.Write([]byte("test"))
	assert.Error(t, err, "Expected error from short write")
	assert.Equal(t, 1, n, "Expected byte count to return from underlying writer")
}

func TestTeeSync_PropagatesErrors(t *testing.T) {
	badsink := &unsyncable{}
	ws := Tee(Discard, badsink)

	assert.Error(t, ws.Sync(), "Expected sync error to propagate")
}

func TestTeeSync_NoErrorsOnDiscard(t *testing.T) {
	ws := Tee(Discard)
	assert.NoError(t, ws.Sync(), "Expected error-free sync to /dev/null")
}

type failingWriter struct {
	err    error
	nBytes int
}

func (f failingWriter) Write(p []byte) (int, error) {
	return f.nBytes, f.err
}

type unsyncable struct {
	bytes.Buffer
}

func (unsyncable) Sync() error { return errors.New("can't sink") }
