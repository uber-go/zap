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
	"time"

	"io"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
	"go.uber.org/zap/internal/ztest"
)

type writeSyncSpy struct {
	io.Writer
	ztest.Syncer
}

type errorWriter struct{}

func (*errorWriter) Write([]byte) (int, error) { return 0, errors.New("unimplemented") }

func requireWriteWorks(t testing.TB, ws WriteSyncer) {
	n, err := ws.Write([]byte("foo"))
	require.NoError(t, err, "Unexpected error writing to WriteSyncer.")
	require.Equal(t, 3, n, "Wrote an unexpected number of bytes.")
}

func TestAddSyncWriteSyncer(t *testing.T) {
	buf := &bytes.Buffer{}
	concrete := &writeSyncSpy{Writer: buf}
	ws := AddSync(concrete)
	requireWriteWorks(t, ws)

	require.NoError(t, ws.Sync(), "Unexpected error syncing a WriteSyncer.")
	require.True(t, concrete.Called(), "Expected to dispatch to concrete type's Sync method.")

	concrete.SetError(errors.New("fail"))
	assert.Error(t, ws.Sync(), "Expected to propagate errors from concrete type's Sync method.")
}

func TestAddSyncWriter(t *testing.T) {
	// If we pass a plain io.Writer, make sure that we still get a WriteSyncer
	// with a no-op Sync.
	buf := &bytes.Buffer{}
	ws := AddSync(buf)
	requireWriteWorks(t, ws)
	assert.NoError(t, ws.Sync(), "Unexpected error calling a no-op Sync method.")
}

func TestBufferWriter(t *testing.T) {
	defer goleak.VerifyNone(t)

	// If we pass a plain io.Writer, make sure that we still get a WriteSyncer
	// with a no-op Sync.
	t.Run("sync", func(t *testing.T) {
		buf := &bytes.Buffer{}
		ws, close := Buffer(AddSync(buf), 0, 0)
		defer close()
		requireWriteWorks(t, ws)
		assert.Empty(t, buf.String(), "Unexpected log calling a no-op Write method.")
		assert.NoError(t, ws.Sync(), "Unexpected error calling a no-op Sync method.")
		assert.Equal(t, "foo", buf.String(), "Unexpected log string")
	})

	t.Run("close", func(t *testing.T) {
		buf := &bytes.Buffer{}
		ws, close := Buffer(AddSync(buf), 0, 0)
		requireWriteWorks(t, ws)
		assert.Empty(t, buf.String(), "Unexpected log calling a no-op Write method.")
		close()
		assert.Equal(t, "foo", buf.String(), "Unexpected log string")
	})

	t.Run("wrap twice", func(t *testing.T) {
		buf := &bytes.Buffer{}
		bufsync, close1 := Buffer(AddSync(buf), 0, 0)
		ws, close2 := Buffer(bufsync, 0, 0)
		requireWriteWorks(t, ws)
		assert.Equal(t, "", buf.String(), "Unexpected log calling a no-op Write method.")
		close2()
		close1()
		assert.Equal(t, "foo", buf.String(), "Unexpected log string")
	})

	t.Run("small buffer", func(t *testing.T) {
		buf := &bytes.Buffer{}
		ws, close := Buffer(AddSync(buf), 5, 0)
		defer close()
		requireWriteWorks(t, ws)
		assert.Equal(t, "", buf.String(), "Unexpected log calling a no-op Write method.")
		requireWriteWorks(t, ws)
		assert.Equal(t, "foo", buf.String(), "Unexpected log string")
	})

	t.Run("flush error", func(t *testing.T) {
		ws, close := Buffer(AddSync(&errorWriter{}), 4, time.Nanosecond)
		n, err := ws.Write([]byte("foo"))
		require.NoError(t, err, "Unexpected error writing to WriteSyncer.")
		require.Equal(t, 3, n, "Wrote an unexpected number of bytes.")
		ws.Write([]byte("foo"))
		assert.Error(t, close(), "Expected close to fail.")
	})

	t.Run("flush timer", func(t *testing.T) {
		buf := &bytes.Buffer{}
		ws, close := Buffer(AddSync(buf), 6, time.Microsecond)
		defer close()
		requireWriteWorks(t, ws)
		ztest.Sleep(10 * time.Millisecond)
		bws := ws.(*bufferWriterSyncer)
		bws.Lock()
		assert.Equal(t, "foo", buf.String(), "Unexpected log string")
		bws.Unlock()

		// flush twice to validate loop logic
		requireWriteWorks(t, ws)
		ztest.Sleep(10 * time.Millisecond)
		bws.Lock()
		assert.Equal(t, "foofoo", buf.String(), "Unexpected log string")
		bws.Unlock()
	})
}

func TestNewMultiWriteSyncerWorksForSingleWriter(t *testing.T) {
	w := &ztest.Buffer{}

	ws := NewMultiWriteSyncer(w)
	assert.Equal(t, w, ws, "Expected NewMultiWriteSyncer to return the same WriteSyncer object for a single argument.")

	ws.Sync()
	assert.True(t, w.Called(), "Expected Sync to be called on the created WriteSyncer")
}

func TestMultiWriteSyncerWritesBoth(t *testing.T) {
	first := &bytes.Buffer{}
	second := &bytes.Buffer{}
	ws := NewMultiWriteSyncer(AddSync(first), AddSync(second))

	msg := []byte("dumbledore")
	n, err := ws.Write(msg)
	require.NoError(t, err, "Expected successful buffer write")
	assert.Equal(t, len(msg), n)

	assert.Equal(t, msg, first.Bytes())
	assert.Equal(t, msg, second.Bytes())
}

func TestMultiWriteSyncerFailsWrite(t *testing.T) {
	ws := NewMultiWriteSyncer(AddSync(&ztest.FailWriter{}))
	_, err := ws.Write([]byte("test"))
	assert.Error(t, err, "Write error should propagate")
}

func TestMultiWriteSyncerFailsShortWrite(t *testing.T) {
	ws := NewMultiWriteSyncer(AddSync(&ztest.ShortWriter{}))
	n, err := ws.Write([]byte("test"))
	assert.NoError(t, err, "Expected fake-success from short write")
	assert.Equal(t, 3, n, "Expected byte count to return from underlying writer")
}

func TestWritestoAllSyncs_EvenIfFirstErrors(t *testing.T) {
	failer := &ztest.FailWriter{}
	second := &bytes.Buffer{}
	ws := NewMultiWriteSyncer(AddSync(failer), AddSync(second))

	_, err := ws.Write([]byte("fail"))
	assert.Error(t, err, "Expected error from call to a writer that failed")
	assert.Equal(t, []byte("fail"), second.Bytes(), "Expected second sink to be written after first error")
}

func TestMultiWriteSyncerSync_PropagatesErrors(t *testing.T) {
	badsink := &ztest.Buffer{}
	badsink.SetError(errors.New("sink is full"))
	ws := NewMultiWriteSyncer(&ztest.Discarder{}, badsink)

	assert.Error(t, ws.Sync(), "Expected sync error to propagate")
}

func TestMultiWriteSyncerSync_NoErrorsOnDiscard(t *testing.T) {
	ws := NewMultiWriteSyncer(&ztest.Discarder{})
	assert.NoError(t, ws.Sync(), "Expected error-free sync to /dev/null")
}

func TestMultiWriteSyncerSync_AllCalled(t *testing.T) {
	failed, second := &ztest.Buffer{}, &ztest.Buffer{}

	failed.SetError(errors.New("disposal broken"))
	ws := NewMultiWriteSyncer(failed, second)

	assert.Error(t, ws.Sync(), "Expected first sink to fail")
	assert.True(t, failed.Called(), "Expected first sink to have Sync method called.")
	assert.True(t, second.Called(), "Expected call to Sync even with first failure.")
}
