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
	"testing"

	"io"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type writeSyncSpy struct {
	io.Writer
	TestSyncer
}

func requireWriteWorks(t testing.TB, ws Pusher) {
	n, err := ws.Push(DebugLevel, []byte("foo"))
	require.NoError(t, err, "Unexpected error writing to Pusher.")
	require.Equal(t, 3, n, "Wrote an unexpected number of bytes.")
}

func TestAddSyncPusher(t *testing.T) {
	buf := &bytes.Buffer{}
	concrete := &writeSyncSpy{Writer: buf}
	ws := IgnoreLevel(AddSync(concrete))
	requireWriteWorks(t, ws)

	require.NoError(t, ws.Sync(), "Unexpected error syncing a Pusher.")
	require.True(t, concrete.Called(), "Expected to dispatch to concrete type's Sync method.")

	concrete.SetError(errors.New("fail"))
	assert.Error(t, ws.Sync(), "Expected to propagate errors from concrete type's Sync method.")
}

func TestAddSyncWriter(t *testing.T) {
	// If we pass a plain io.Writer, make sure that we still get a Pusher
	// with a no-op Sync.
	buf := &bytes.Buffer{}
	ws := IgnoreLevel(AddSync(buf))
	requireWriteWorks(t, ws)
	assert.NoError(t, ws.Sync(), "Unexpected error calling a no-op Sync method.")
}

func TestMultiPusherWritesBoth(t *testing.T) {
	first := &bytes.Buffer{}
	second := &bytes.Buffer{}
	ws := NewMultiPusher(IgnoreLevel(AddSync(first)), IgnoreLevel(AddSync(second)))

	msg := []byte("dumbledore")
	n, err := ws.Push(DebugLevel, msg)
	require.NoError(t, err, "Expected successful buffer write")
	assert.Equal(t, len(msg), n)

	assert.Equal(t, msg, first.Bytes())
	assert.Equal(t, msg, second.Bytes())
}

func TestMultiPusherFailsWrite(t *testing.T) {
	ws := NewMultiPusher(&TestFailPusher{})
	_, err := ws.Push(DebugLevel, []byte("test"))
	assert.Error(t, err, "Write error should propagate")
}

func TestMultiPusherFailsShortWrite(t *testing.T) {
	ws := NewMultiPusher(&TestShortPusher{})
	n, err := ws.Push(DebugLevel, []byte("test"))
	assert.NoError(t, err, "Expected fake-success from short write")
	assert.Equal(t, 3, n, "Expected byte count to return from underlying writer")
}

func TestWritestoAllSyncs_EvenIfFirstErrors(t *testing.T) {
	failer := &TestFailPusher{}
	second := &bytes.Buffer{}
	ws := NewMultiPusher(failer, IgnoreLevel(AddSync(second)))

	_, err := ws.Push(DebugLevel, []byte("fail"))
	assert.Error(t, err, "Expected error from call to a writer that failed")
	assert.Equal(t, []byte("fail"), second.Bytes(), "Expected second sink to be written after first error")
}

func TestMultiPusherSync_PropagatesErrors(t *testing.T) {
	badsink := &TestBuffer{}
	badsink.SetError(errors.New("sink is full"))
	ws := NewMultiPusher(&TestDiscarder{}, badsink)

	assert.Error(t, ws.Sync(), "Expected sync error to propagate")
}

func TestMultiPusherSync_NoErrorsOnDiscard(t *testing.T) {
	ws := NewMultiPusher(&TestDiscarder{})
	assert.NoError(t, ws.Sync(), "Expected error-free sync to /dev/null")
}

func TestMultiPusherSync_AllCalled(t *testing.T) {
	failed, second := &TestBuffer{}, &TestBuffer{}

	failed.SetError(errors.New("disposal broken"))
	ws := NewMultiPusher(failed, second)

	assert.Error(t, ws.Sync(), "Expected first sink to fail")
	assert.True(t, failed.Called(), "Expected first sink to have Sync method called.")
	assert.True(t, second.Called(), "Expected call to Sync even with first failure.")
}
