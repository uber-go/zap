// Copyright (c) 2022 Uber Technologies, Inc.
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

//go:build windows

package zap

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWindowsPaths(t *testing.T) {
	// See https://docs.microsoft.com/en-us/dotnet/standard/io/file-path-formats
	tests := []struct {
		msg  string
		path string
	}{
		{
			msg:  "local path with drive",
			path: `c:\log.json`,
		},
		{
			msg:  "local path with drive using forward slash",
			path: `c:/log.json`,
		},
		{
			msg:  "local path without drive",
			path: `\Temp\log.json`,
		},
		{
			msg:  "unc path",
			path: `\\Server2\Logs\log.json`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.msg, func(t *testing.T) {
			sr := newSinkRegistry()

			openFilename := "<not called>"
			sr.openFile = func(filename string, _ int, _ os.FileMode) (*os.File, error) {
				openFilename = filename
				return nil, assert.AnError
			}

			_, err := sr.newSink(tt.path)
			assert.Equal(t, assert.AnError, err, "expect stub error from OpenFile")
			assert.Equal(t, tt.path, openFilename, "unexpected path opened")
		})
	}
}
