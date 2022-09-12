// Copyright (c) 2016-2022 Uber Technologies, Inc.
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
	"errors"
	"io"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/multierr"
	"go.uber.org/zap/zapcore"
)

func TestOpenNoPaths(t *testing.T) {
	ws, cleanup, err := Open()
	defer cleanup()

	assert.NoError(t, err, "Expected opening no paths to succeed.")
	assert.Equal(
		t,
		zapcore.AddSync(io.Discard),
		ws,
		"Expected opening no paths to return a no-op WriteSyncer.",
	)
}

func TestOpen(t *testing.T) {
	tempName := filepath.Join(t.TempDir(), "test.log")
	assert.False(t, fileExists(tempName))
	require.True(t, filepath.IsAbs(tempName), "Expected absolute temp file path.")

	tests := []struct {
		msg   string
		paths []string
	}{
		{
			msg:   "stdout",
			paths: []string{"stdout"},
		},
		{
			msg:   "stderr",
			paths: []string{"stderr"},
		},
		{
			msg:   "temp file path only",
			paths: []string{tempName},
		},
		{
			msg:   "temp file file scheme",
			paths: []string{"file://" + tempName},
		},
		{
			msg:   "temp file with file scheme and host localhost",
			paths: []string{"file://localhost" + tempName},
		},
	}

	for _, tt := range tests {
		t.Run(tt.msg, func(t *testing.T) {
			_, cleanup, err := Open(tt.paths...)
			if err == nil {
				defer cleanup()
			}

			assert.NoError(t, err, "Unexpected error opening paths %v.", tt.paths)
		})
	}

	assert.True(t, fileExists(tempName))
	os.Remove(tempName)
}

func TestOpenPathsNotFound(t *testing.T) {
	tempName := filepath.Join(t.TempDir(), "test.log")

	tests := []struct {
		msg               string
		paths             []string
		wantNotFoundPaths []string
	}{
		{
			msg:               "missing path",
			paths:             []string{"/foo/bar/baz"},
			wantNotFoundPaths: []string{"/foo/bar/baz"},
		},
		{
			msg:               "missing file scheme url with host localhost",
			paths:             []string{"file://localhost/foo/bar/baz"},
			wantNotFoundPaths: []string{"/foo/bar/baz"},
		},
		{
			msg:   "multiple paths",
			paths: []string{"stdout", "/foo/bar/baz", tempName, "file:///baz/quux"},
			wantNotFoundPaths: []string{
				"/foo/bar/baz",
				"/baz/quux",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.msg, func(t *testing.T) {
			_, cleanup, err := Open(tt.paths...)
			if !assert.Error(t, err, "Open must fail.") {
				cleanup()
				return
			}

			errs := multierr.Errors(err)
			require.Len(t, errs, len(tt.wantNotFoundPaths))
			for i, err := range errs {
				assert.ErrorIs(t, err, fs.ErrNotExist)
				assert.ErrorContains(t, err, tt.wantNotFoundPaths[i], "missing path in error")
			}
		})
	}
}

func TestOpenRelativePath(t *testing.T) {
	const name = "test-relative-path.txt"

	require.False(t, fileExists(name), "Test file already exists.")
	s, cleanup, err := Open(name)
	require.NoError(t, err, "Open failed.")
	defer func() {
		err := os.Remove(name)
		if !t.Failed() {
			// If the test has already failed, we probably didn't create this file.
			require.NoError(t, err, "Deleting test file failed.")
		}
	}()
	defer cleanup()

	_, err = s.Write([]byte("test"))
	assert.NoError(t, err, "Write failed.")
	assert.True(t, fileExists(name), "Didn't create file for relative path.")
}

func TestOpenFails(t *testing.T) {
	tests := []struct {
		paths []string
	}{
		{paths: []string{"./non-existent-dir/file"}},           // directory doesn't exist
		{paths: []string{"stdout", "./non-existent-dir/file"}}, // directory doesn't exist
		{paths: []string{"://foo.log"}},                        // invalid URL, scheme can't begin with colon
		{paths: []string{"mem://somewhere"}},                   // scheme not registered
	}

	for _, tt := range tests {
		_, cleanup, err := Open(tt.paths...)
		require.Nil(t, cleanup, "Cleanup function should never be nil")
		assert.Error(t, err, "Open with invalid URL should fail.")
	}
}

func TestOpenOtherErrors(t *testing.T) {
	tempName := filepath.Join(t.TempDir(), "test.log")

	tests := []struct {
		msg     string
		paths   []string
		wantErr string
	}{
		{
			msg:     "file with unexpected host",
			paths:   []string{"file://host01.test.com" + tempName},
			wantErr: "empty or use localhost",
		},
		{
			msg:     "file with user on localhost",
			paths:   []string{"file://rms@localhost" + tempName},
			wantErr: "user and password not allowed",
		},
		{
			msg:     "file url with fragment",
			paths:   []string{"file://localhost" + tempName + "#foo"},
			wantErr: "fragments not allowed",
		},
		{
			msg:     "file url with query",
			paths:   []string{"file://localhost" + tempName + "?foo=bar"},
			wantErr: "query parameters not allowed",
		},
		{
			msg:     "file with port",
			paths:   []string{"file://localhost:8080" + tempName},
			wantErr: "ports not allowed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.msg, func(t *testing.T) {
			_, cleanup, err := Open(tt.paths...)
			if !assert.Error(t, err, "Open must fail.") {
				cleanup()
				return
			}

			assert.ErrorContains(t, err, tt.wantErr, "Unexpected error opening paths %v.", tt.paths)
		})
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

func TestOpenWithErroringSinkFactory(t *testing.T) {
	stubSinkRegistry(t)

	msg := "expected factory error"
	factory := func(_ *url.URL) (Sink, error) {
		return nil, errors.New(msg)
	}

	assert.NoError(t, RegisterSink("test", factory), "Failed to register sink factory.")
	_, _, err := Open("test://some/path")
	assert.ErrorContains(t, err, msg)
}

func TestCombineWriteSyncers(t *testing.T) {
	tw := &testWriter{"test", t}
	w := CombineWriteSyncers(tw)
	w.Write([]byte("test"))
}

func fileExists(name string) bool {
	if _, err := os.Stat(name); os.IsNotExist(err) {
		return false
	}
	return true
}
