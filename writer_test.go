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
	"io/ioutil"
	"os"
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
	temp, err := ioutil.TempFile("", "zap-open-test")
	require.NoError(t, err, "Couldn't create a temporary file for test.")
	defer os.Remove(temp.Name())

	tests := []struct {
		paths     []string
		filenames []string
		error     string
	}{
		{[]string{"stdout"}, []string{os.Stdout.Name()}, ""},
		{[]string{"stderr"}, []string{os.Stderr.Name()}, ""},
		{[]string{temp.Name()}, []string{temp.Name()}, ""},
		{[]string{"/foo/bar/baz"}, []string{}, "open /foo/bar/baz: no such file or directory"},
		{
			paths:     []string{"stdout", "/foo/bar/baz", temp.Name(), "/baz/quux"},
			filenames: []string{os.Stdout.Name(), temp.Name()},
			error:     "open /foo/bar/baz: no such file or directory; open /baz/quux: no such file or directory",
		},
	}

	for _, tt := range tests {
		wss, cleanup, err := open(tt.paths)
		if err == nil {
			defer cleanup()
		}

		if tt.error == "" {
			assert.NoError(t, err, "Unexpected error opening paths %v.", tt.paths)
		} else {
			assert.Equal(t, tt.error, err.Error(), "Unexpected error opening paths %v.", tt.paths)
		}
		names := make([]string, len(wss))
		for i, ws := range wss {
			f, ok := ws.(*os.File)
			require.True(t, ok, "Expected all WriteSyncers returned from open() to be files.")
			names[i] = f.Name()
		}
		assert.Equal(t, tt.filenames, names, "Opened unexpected files given paths %v.", tt.paths)
	}
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

func TestCombineWriteSyncers(t *testing.T) {
	tw := &testWriter{"test", t}
	w := CombineWriteSyncers(tw)
	w.Write([]byte("test"))
}
