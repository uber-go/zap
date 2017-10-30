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

package zaptest

import (
	"bufio"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func readCode(t testing.TB, fname string) []string {
	f, err := os.Open(fname)
	require.NoError(t, err, "Failed to read %s.", fname)
	defer func() {
		require.NoError(t, f.Close(), "Error closing file %s.", fname)
	}()

	var lines []string
	s := bufio.NewScanner(f)
	for s.Scan() {
		l := s.Text()
		if len(l) == 0 {
			continue
		}
		if strings.HasPrefix(l, "//") {
			continue
		}
		if strings.HasPrefix(l, "package ") {
			continue
		}
		lines = append(lines, l)
	}
	return lines
}

func TestCopiedCodeInSync(t *testing.T) {
	// Until we drop Go 1.8 support, we need to keep a near-exact copy of the
	// ztest package's WriteSyncer test spies in zaptest. This test ensures that
	// the two files stay in sync.
	assert.Equal(t,
		readCode(t, "../internal/ztest/writer.go"),
		readCode(t, "writer.go"),
		"Writer spy implementations in zaptest and internal/ztest should be identical.",
	)
}

func TestSyncer(t *testing.T) {
	err := errors.New("sentinel")
	s := &Syncer{}
	s.SetError(err)
	assert.Equal(t, err, s.Sync(), "Expected Sync to fail with provided error.")
	assert.True(t, s.Called(), "Expected to record that Sync was called.")
}

func TestDiscarder(t *testing.T) {
	d := &Discarder{}
	payload := []byte("foo")
	n, err := d.Write(payload)
	assert.NoError(t, err, "Unexpected error writing to Discarder.")
	assert.Equal(t, len(payload), n, "Wrong number of bytes written.")
}

func TestFailWriter(t *testing.T) {
	w := &FailWriter{}
	payload := []byte("foo")
	n, err := w.Write(payload)
	assert.Error(t, err, "Expected an error writing to FailWriter.")
	assert.Equal(t, len(payload), n, "Wrong number of bytes written.")
}

func TestShortWriter(t *testing.T) {
	w := &ShortWriter{}
	payload := []byte("foo")
	n, err := w.Write(payload)
	assert.NoError(t, err, "Unexpected error writing to ShortWriter.")
	assert.Equal(t, len(payload)-1, n, "Wrong number of bytes written.")
}

func TestBuffer(t *testing.T) {
	buf := &Buffer{}
	buf.WriteString("foo\n")
	buf.WriteString("bar\n")
	assert.Equal(t, []string{"foo", "bar"}, buf.Lines(), "Unexpected output from Lines.")
	assert.Equal(t, "foo\nbar", buf.Stripped(), "Unexpected output from Stripped.")
}
