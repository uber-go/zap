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

package zapcore_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"go.uber.org/zap/testutils"
	. "go.uber.org/zap/zapcore"
)

func TestBufferedWriterPropagatesFailures(t *testing.T) {
	failSyncer := &testutils.Buffer{}
	failSyncer.SetError(errors.New("no sync"))

	failWriteSyncer := &testutils.FailWriter{}
	failWriteSyncer.SetError(errors.New("no sync"))

	tests := []struct {
		desc           string
		initial        string
		try            string
		ws             WriteSyncer
		n              int
		expectWriteErr string
		expectSyncErr  string
	}{
		{
			desc: "no write, no buffered content, fail writes",
			ws:   &testutils.FailWriter{},
		},
		{
			desc:          "no write, some buffered content, fail writes",
			initial:       "ab",
			ws:            &testutils.FailWriter{},
			expectSyncErr: "failed",
		},
		{
			desc:           "write overflows buffer, fail writes",
			initial:        "ab",
			try:            "cde",
			ws:             &testutils.FailWriter{},
			n:              3,
			expectWriteErr: "failed",
			expectSyncErr:  "failed",
		},
		{
			desc:           "write can't be buffered, fail writes",
			initial:        "ab",
			try:            "cdefg",
			ws:             &testutils.FailWriter{},
			n:              5,
			expectWriteErr: "failed; failed",
		},
		{
			desc:          "no write, no buffered content, fail syncs",
			ws:            failSyncer,
			expectSyncErr: "no sync",
		},
		{
			desc:           "write overflows buffer, fail writes and syncs",
			initial:        "ab",
			try:            "cde",
			ws:             failWriteSyncer,
			n:              3,
			expectWriteErr: "failed",
			expectSyncErr:  "failed; no sync",
		},
	}

	for _, tt := range tests {
		buf := Buffer(4, tt.ws)
		_, err := buf.Write([]byte(tt.initial))
		assert.NoError(t, err, "Unexpected error writing initial buffer contents (%q).", tt.desc)

		n, err := buf.Write([]byte(tt.try))
		assert.Equal(t, tt.n, n, "Unexpected number of bytes written (%q).", tt.desc)

		if tt.expectWriteErr == "" {
			assert.NoError(t, err, "Unexpected error calling Write (%q).", tt.desc)
		} else {
			assert.Equal(t, tt.expectWriteErr, err.Error(), "Write error didn't match expectations (%q).", tt.desc)
		}

		if tt.expectSyncErr == "" {
			assert.NoError(t, buf.Sync(), "Unexpected error calling Sync (%q).", tt.desc)
		} else {
			assert.Equal(t, tt.expectSyncErr, buf.Sync().Error(), "Sync error didn't match expectations (%q).", tt.desc)
		}
	}
}
