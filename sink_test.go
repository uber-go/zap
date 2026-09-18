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
	"io"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.uber.org/zap/zapcore"
)

func stubSinkRegistry(t testing.TB) *sinkRegistry {
	origSinkRegistry := _sinkRegistry
	t.Cleanup(func() {
		_sinkRegistry = origSinkRegistry
	})

	r := newSinkRegistry()
	_sinkRegistry = r
	return r
}

func TestRegisterSink(t *testing.T) {
	stubSinkRegistry(t)

	const (
		memScheme = "mem"
		nopScheme = "no-op.1234"
	)
	var memCalls, nopCalls int

	buf := bytes.NewBuffer(nil)
	memFactory := func(u *url.URL) (Sink, error) {
		assert.Equal(t, u.Scheme, memScheme, "Scheme didn't match registration.")
		memCalls++
		return nopCloserSink{zapcore.AddSync(buf)}, nil
	}
	nopFactory := func(u *url.URL) (Sink, error) {
		assert.Equal(t, u.Scheme, nopScheme, "Scheme didn't match registration.")
		nopCalls++
		return nopCloserSink{zapcore.AddSync(io.Discard)}, nil
	}

	require.NoError(t, RegisterSink(strings.ToUpper(memScheme), memFactory), "Failed to register scheme %q.", memScheme)
	require.NoError(t, RegisterSink(nopScheme, nopFactory), "Failed to register scheme %q.", nopScheme)

	sink, closeSink, err := Open(
		memScheme+"://somewhere",
		nopScheme+"://somewhere-else",
	)
	require.NoError(t, err, "Unexpected error opening URLs with registered schemes.")
	defer closeSink()

	assert.Equal(t, 1, memCalls, "Unexpected number of calls to memory factory.")
	assert.Equal(t, 1, nopCalls, "Unexpected number of calls to no-op factory.")

	_, err = sink.Write([]byte("foo"))
	assert.NoError(t, err, "Failed to write to combined WriteSyncer.")
	assert.Equal(t, "foo", buf.String(), "Unexpected buffer contents.")
}

func TestRegisterSinkErrors(t *testing.T) {
	nopFactory := func(_ *url.URL) (Sink, error) {
		return nopCloserSink{zapcore.AddSync(io.Discard)}, nil
	}
	tests := []struct {
		scheme string
		err    string
	}{
		{"", "empty string"},
		{"FILE", "already registered"},
		{"42", "not a valid scheme"},
		{"http*", "not a valid scheme"},
	}

	for _, tt := range tests {
		t.Run("scheme-"+tt.scheme, func(t *testing.T) {
			r := newSinkRegistry()
			err := r.RegisterSink(tt.scheme, nopFactory)
			assert.ErrorContains(t, err, tt.err)
		})
	}
}

func TestFileURLPath(t *testing.T) {
	tests := []struct {
		msg       string
		path      string
		isWindows bool
		want      string
	}{
		{msg: "windows drive letter", path: "/C:/logs/app.log", isWindows: true, want: "C:/logs/app.log"},
		{msg: "windows lowercase drive letter", path: "/d:/app.log", isWindows: true, want: "d:/app.log"},
		{msg: "windows path without a drive", path: "/logs/app.log", isWindows: true, want: "/logs/app.log"},
		{msg: "windows bare drive", path: "/C:", isWindows: true, want: "/C:"},
		{msg: "windows non-letter drive", path: "/1:/app.log", isWindows: true, want: "/1:/app.log"},
		{msg: "windows empty path", path: "", isWindows: true, want: ""},
		{msg: "non-windows keeps the slash", path: "/C:/logs/app.log", isWindows: false, want: "/C:/logs/app.log"},
		{msg: "non-windows absolute path", path: "/var/log/app.log", isWindows: false, want: "/var/log/app.log"},
	}

	for _, tt := range tests {
		t.Run(tt.msg, func(t *testing.T) {
			sr := newSinkRegistry()
			sr.isWindows = tt.isWindows
			assert.Equal(t, tt.want, sr.fileURLPath(tt.path))
		})
	}
}

func TestWindowsFileURLs(t *testing.T) {
	// RFC 8089 file URLs for Windows paths, see
	// https://github.com/uber-go/zap/issues/621. These run on every platform
	// by forcing the registry into Windows mode and stubbing the file open.
	tests := []struct {
		msg      string
		url      string
		wantPath string
	}{
		{msg: "empty host", url: "file:///C:/logs/app.log", wantPath: "C:/logs/app.log"},
		{msg: "localhost", url: "file://localhost/C:/logs/app.log", wantPath: "C:/logs/app.log"},
		{msg: "percent-encoded space", url: "file:///C:/My%20Logs/app.log", wantPath: "C:/My Logs/app.log"},
	}

	for _, tt := range tests {
		t.Run(tt.msg, func(t *testing.T) {
			sr := newSinkRegistry()
			sr.isWindows = true

			openFilename := "<not called>"
			sr.openFile = func(filename string, _ int, _ os.FileMode) (*os.File, error) {
				openFilename = filename
				return nil, assert.AnError
			}

			_, err := sr.newSink(tt.url)
			assert.Equal(t, assert.AnError, err, "expect stub error from OpenFile")
			assert.Equal(t, tt.wantPath, openFilename, "unexpected path opened")
		})
	}
}
