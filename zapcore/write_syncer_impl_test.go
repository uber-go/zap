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
	"go.uber.org/zap/testutils"
)

type writeSyncSpy struct {
	io.Writer
	testutils.Syncer
}

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

func TestBufferedWriterWrite(t *testing.T) {
	tests := []struct {
		desc    string
		wrap    func(WriteSyncer) WriteSyncer // wrap or replace the underlying WriteSyncer
		initial string                        // initial buffer contents
		try     string                        // bytes to write
		buf     string                        // final buffer contents
		flushed string                        // written to underlying WriteSyncer
		n       int
		err     error
	}{
		{
			desc:    "new write fills buffer",
			initial: "ab",
			try:     "cd",
			buf:     "abcd",
			flushed: "",
			n:       2,
		},
		{
			desc:    "new write triggers buffer flush",
			initial: "ab",
			try:     "cde",
			buf:     "cde",
			flushed: "ab",
			n:       3,
		},
		{
			desc:    "new write is too big to buffer",
			initial: "ab",
			try:     "cdefg",
			buf:     "",
			flushed: "abcdefg",
			n:       5,
		},
	}

	for _, tt := range tests {
		sink := &bytes.Buffer{}
		underlying := AddSync(sink)

		b := Buffer(4, underlying).(*bufferedWriter)
		b.buf = []byte(tt.initial)

		n, err := b.Write([]byte(tt.try))
		assert.Equal(t, tt.n, n, "Unexpected number of bytes written (%s).", tt.desc)
		assert.Equal(t, tt.err, err, "Unexpected error (%s).", tt.desc)
		assert.Equal(t, tt.buf, string(b.buf), "Unexpected buffer contents (%s).", tt.desc)
		assert.Equal(t, tt.flushed, sink.String(), "Unexpected bytes written to underlying WriteSyncer (%s).", tt.desc)
	}
}

func TestBufferedWriterSync(t *testing.T) {
	sink := &bytes.Buffer{}
	underlying := AddSync(sink)

	b := Buffer(4, underlying).(*bufferedWriter)
	b.buf = []byte("abc")

	assert.NoError(t, b.Sync(), "Unexpected error syncing.")
	assert.Equal(t, "", string(b.buf), "Expected to write out buffer contents during Sync.")
	assert.Equal(t, "abc", sink.String(), "Unexpected bytes written to underlying WriteSyncer.")
}

func TestMultiWriteSyncerWritesBoth(t *testing.T) {
	first := &bytes.Buffer{}
	second := &bytes.Buffer{}
	ws := MultiWriteSyncer(AddSync(first), AddSync(second))

	msg := []byte("dumbledore")
	n, err := ws.Write(msg)
	require.NoError(t, err, "Expected successful buffer write")
	assert.Equal(t, len(msg), n)

	assert.Equal(t, msg, first.Bytes())
	assert.Equal(t, msg, second.Bytes())
}

func TestMultiWriteSyncerFailsWrite(t *testing.T) {
	ws := MultiWriteSyncer(AddSync(&testutils.FailWriter{}))
	_, err := ws.Write([]byte("test"))
	assert.Error(t, err, "Write error should propagate")
}

func TestMultiWriteSyncerFailsShortWrite(t *testing.T) {
	ws := MultiWriteSyncer(AddSync(&testutils.ShortWriter{}))
	n, err := ws.Write([]byte("test"))
	assert.NoError(t, err, "Expected fake-success from short write")
	assert.Equal(t, 3, n, "Expected byte count to return from underlying writer")
}

func TestWritestoAllSyncs_EvenIfFirstErrors(t *testing.T) {
	failer := &testutils.FailWriter{}
	second := &bytes.Buffer{}
	ws := MultiWriteSyncer(AddSync(failer), AddSync(second))

	_, err := ws.Write([]byte("fail"))
	assert.Error(t, err, "Expected error from call to a writer that failed")
	assert.Equal(t, []byte("fail"), second.Bytes(), "Expected second sink to be written after first error")
}

func TestMultiWriteSyncerSync_PropagatesErrors(t *testing.T) {
	badsink := &testutils.Buffer{}
	badsink.SetError(errors.New("sink is full"))
	ws := MultiWriteSyncer(&testutils.Discarder{}, badsink)

	assert.Error(t, ws.Sync(), "Expected sync error to propagate")
}

func TestMultiWriteSyncerSync_NoErrorsOnDiscard(t *testing.T) {
	ws := MultiWriteSyncer(&testutils.Discarder{})
	assert.NoError(t, ws.Sync(), "Expected error-free sync to /dev/null")
}

func TestMultiWriteSyncerSync_AllCalled(t *testing.T) {
	failed, second := &testutils.Buffer{}, &testutils.Buffer{}

	failed.SetError(errors.New("disposal broken"))
	ws := MultiWriteSyncer(failed, second)

	assert.Error(t, ws.Sync(), "Expected first sink to fail")
	assert.True(t, failed.Called(), "Expected first sink to have Sync method called.")
	assert.True(t, second.Called(), "Expected call to Sync even with first failure.")
}
