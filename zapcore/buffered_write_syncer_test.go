// Copyright (c) 2021 Uber Technologies, Inc.
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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/internal/ztest"
)

func TestBufferWriter(t *testing.T) {
	// If we pass a plain io.Writer, make sure that we still get a WriteSyncer
	// with a no-op Sync.
	t.Run("sync", func(t *testing.T) {
		buf := &bytes.Buffer{}
		ws := &BufferedWriteSyncer{WriteSyncer: AddSync(buf)}

		requireWriteWorks(t, ws)
		assert.Empty(t, buf.String(), "Unexpected log calling a no-op Write method.")
		assert.NoError(t, ws.Sync(), "Unexpected error calling a no-op Sync method.")
		assert.Equal(t, "foo", buf.String(), "Unexpected log string")
		assert.NoError(t, ws.Close())
	})

	t.Run("close", func(t *testing.T) {
		buf := &bytes.Buffer{}
		ws := &BufferedWriteSyncer{WriteSyncer: AddSync(buf)}
		requireWriteWorks(t, ws)
		assert.Empty(t, buf.String(), "Unexpected log calling a no-op Write method.")
		assert.NoError(t, ws.Close())
		assert.Equal(t, "foo", buf.String(), "Unexpected log string")
	})

	t.Run("wrap twice", func(t *testing.T) {
		buf := &bytes.Buffer{}
		bufsync := &BufferedWriteSyncer{WriteSyncer: AddSync(buf)}
		ws := &BufferedWriteSyncer{WriteSyncer: bufsync}
		requireWriteWorks(t, ws)
		assert.Equal(t, "", buf.String(), "Unexpected log calling a no-op Write method.")
		assert.NoError(t, ws.Close())
		assert.NoError(t, bufsync.Close())
		assert.Equal(t, "foo", buf.String(), "Unexpected log string")
	})

	t.Run("small buffer", func(t *testing.T) {
		buf := &bytes.Buffer{}
		ws := &BufferedWriteSyncer{WriteSyncer: AddSync(buf), Size: 5}

		requireWriteWorks(t, ws)
		assert.Equal(t, "", buf.String(), "Unexpected log calling a no-op Write method.")
		requireWriteWorks(t, ws)
		assert.Equal(t, "foo", buf.String(), "Unexpected log string")
		assert.NoError(t, ws.Close())
	})

	t.Run("flush error", func(t *testing.T) {
		ws := &BufferedWriteSyncer{WriteSyncer: &ztest.FailWriter{}, Size: 4}
		n, err := ws.Write([]byte("foo"))
		require.NoError(t, err, "Unexpected error writing to WriteSyncer.")
		require.Equal(t, 3, n, "Wrote an unexpected number of bytes.")
		ws.Write([]byte("foo"))
		assert.Error(t, ws.Close(), "Expected close to fail.")
	})

	t.Run("flush timer", func(t *testing.T) {
		buf := &bytes.Buffer{}
		clock := newControlledClock()
		ws := &BufferedWriteSyncer{
			WriteSyncer:   AddSync(buf),
			Size:          6,
			FlushInterval: time.Microsecond,
			Clock:         clock,
		}
		requireWriteWorks(t, ws)
		clock.Add(10 * time.Millisecond)
		assert.Equal(t, "foo", buf.String(), "Unexpected log string")

		// flush twice to validate loop logic
		requireWriteWorks(t, ws)
		clock.Add(10 * time.Millisecond)
		assert.Equal(t, "foofoo", buf.String(), "Unexpected log string")
		assert.NoError(t, ws.Close())
	})
}
