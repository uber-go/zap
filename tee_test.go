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

	"github.com/uber-go/zap/spywrite"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
	failer := spywrite.FailWriter{}
	ws := MultiWriteSyncer(AddSync(failer))

	_, err := ws.Write([]byte("test"))
	assert.Error(t, err, "Write error should propagate")
}

func TestMultiWriteSyncerFailsShortWrite(t *testing.T) {
	shorter := spywrite.ShortWriter{}
	ws := MultiWriteSyncer(AddSync(shorter))

	n, err := ws.Write([]byte("test"))
	assert.NoError(t, err, "Expected fake-success from short write")
	assert.Equal(t, 3, n, "Expected byte count to return from underlying writer")
}

func TestWritestoAllSyncs_EvenIfFirstErrors(t *testing.T) {
	failer := spywrite.FailWriter{}
	second := &bytes.Buffer{}
	ws := MultiWriteSyncer(AddSync(failer), AddSync(second))

	_, err := ws.Write([]byte("fail"))
	assert.Error(t, err, "Expected error from call to a writer that failed")
	assert.Equal(t, []byte("fail"), second.Bytes(), "Expected second sink to be written after first error")
}

func TestMultiWriteSyncerSync_PropagatesErrors(t *testing.T) {
	badsink := &syncSpy{}
	badsink.SetError(errors.New("sink is full"))
	ws := MultiWriteSyncer(Discard, badsink)

	assert.Error(t, ws.Sync(), "Expected sync error to propagate")
}

func TestMultiWriteSyncerSync_NoErrorsOnDiscard(t *testing.T) {
	ws := MultiWriteSyncer(Discard)
	assert.NoError(t, ws.Sync(), "Expected error-free sync to /dev/null")
}

func TestMultiError_WrapsStrings(t *testing.T) {
	err := multiError{errors.New("battlestar"), errors.New("galactaca")}
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "battlestar")
	assert.Contains(t, err.Error(), "galactaca")
}

func TestMultiWriteSyncerSync_AllCalled(t *testing.T) {
	failedsink := &syncSpy{}
	second := &syncSpy{}

	failedsink.SetError(errors.New("disposal broken"))
	ws := MultiWriteSyncer(failedsink, second)

	assert.Error(t, ws.Sync(), "Expected first sink to fail")
	assert.True(t, second.Called(), "Expected call even with first failure")
}

type syncSpy struct {
	bytes.Buffer
	spywrite.Syncer
}
