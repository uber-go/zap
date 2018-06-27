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
	"encoding/hex"
	"errors"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
)

func TestOpenNoPaths(t *testing.T) {
	ws, cleanup, err := Open()
	defer cleanup()

	assert.NoError(t, err, "Expected opening no paths to succeed.")
	assert.Equal(
		t,
		zapcore.AddSync(ioutil.Discard),
		ws,
		"Expected opening no paths to return a no-op WriteSyncer.",
	)
}

func TestOpen(t *testing.T) {
	tempName := tempFileName("", "zap-open-test")
	assert.False(t, fileExists(tempName))

	tests := []struct {
		paths []string
		error string
	}{
		{[]string{"stdout"}, ""},
		{[]string{"stderr"}, ""},
		{[]string{tempName}, ""},
		{[]string{"/foo/bar/baz"}, "open /foo/bar/baz: no such file or directory"},
		{
			paths: []string{"stdout", "/foo/bar/baz", tempName, "/baz/quux"},
			error: "open /foo/bar/baz: no such file or directory; open /baz/quux: no such file or directory",
		},
	}

	for _, tt := range tests {
		_, cleanup, err := Open(tt.paths...)
		if err == nil {
			defer cleanup()
		}

		if tt.error == "" {
			assert.NoError(t, err, "Unexpected error opening paths %v.", tt.paths)
		} else {
			assert.Equal(t, tt.error, err.Error(), "Unexpected error opening paths %v.", tt.paths)
		}
	}

	assert.True(t, fileExists(tempName))
	os.Remove(tempName)
}

func TestOpenFails(t *testing.T) {
	tests := []struct {
		paths []string
	}{
		{
			paths: []string{"./non-existent-dir/file"},
		},
		{
			paths: []string{"stdout", "./non-existent-dir/file"},
		},
	}

	for _, tt := range tests {
		_, cleanup, err := Open(tt.paths...)
		require.Nil(t, cleanup, "Cleanup function should never be nil")
		assert.Error(t, err, "Open with non-existent directory should fail")
	}
}

type testWriter struct {
	expected string
	t        testing.TB
}

func (w *testWriter) Write(actual []byte) (int, error) {
	assert.Equal(w.t, []byte(w.expected), actual, "Unexpected write error.")
	return len(actual), nil
}

func (w *testWriter) Sync() error {
	return nil
}

func TestOpenWithCustomSink(t *testing.T) {
	defer resetSinkRegistry()
	tw := &testWriter{"test", t}
	ctr := func() (Sink, error) { return nopCloserSink{tw}, nil }
	assert.Nil(t, RegisterSink("TestOpenWithCustomSink", ctr))
	w, cleanup, err := Open("TestOpenWithCustomSink")
	assert.Nil(t, err)
	defer cleanup()
	w.Write([]byte("test"))
}

func TestOpenWithErroringSinkFactory(t *testing.T) {
	defer resetSinkRegistry()
	expectedErr := errors.New("expected factory error")
	ctr := func() (Sink, error) { return nil, expectedErr }
	assert.Nil(t, RegisterSink("TestOpenWithErroringSinkFactory", ctr))
	_, _, err := Open("TestOpenWithErroringSinkFactory")
	assert.Equal(t, expectedErr, err)
}

func TestCombineWriteSyncers(t *testing.T) {
	tw := &testWriter{"test", t}
	w := CombineWriteSyncers(tw)
	w.Write([]byte("test"))
}

func tempFileName(prefix, suffix string) string {
	randBytes := make([]byte, 16)
	rand.Read(randBytes)
	return filepath.Join(os.TempDir(), prefix+hex.EncodeToString(randBytes)+suffix)
}

func fileExists(name string) bool {
	if _, err := os.Stat(name); os.IsNotExist(err) {
		return false
	}
	return true
}
