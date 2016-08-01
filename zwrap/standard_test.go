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

package zwrap

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/uber-go/zap"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newStd(lvl zap.Level) (StandardLogger, *bytes.Buffer, error) {
	buf := &bytes.Buffer{}
	logger := zap.New(
		zap.NewJSONEncoder(),
		zap.DebugLevel,
		zap.Output(zap.AddSync(buf)),
	)
	std, err := Standardize(logger, lvl)
	return std, buf, err
}

func TestStandardizeInvalidLevels(t *testing.T) {
	for _, level := range []zap.Level{zap.PanicLevel, zap.FatalLevel, zap.Level(42)} {
		_, _, err := newStd(level)
		assert.Equal(t, ErrInvalidLevel, err, "Expected ErrInvalidLevel when passing an invalid level to Standardize.")
	}
}

func TestStandardizeValidLevels(t *testing.T) {
	for _, level := range []zap.Level{zap.DebugLevel, zap.InfoLevel, zap.WarnLevel, zap.ErrorLevel} {
		std, buf, err := newStd(level)
		require.NoError(t, err, "Unexpected error calling Standardize with a valid level.")
		std.Print("foo")
		expectation := fmt.Sprintf(`"level":"%s"`, level.String())
		require.Contains(t, buf.String(), expectation, "Print logged at an unexpected level.")
		buf.Reset()
	}
}

func TestStandardLoggerPrint(t *testing.T) {
	std, buf, err := newStd(zap.InfoLevel)
	require.NoError(t, err, "Unexpected error standardizing a Logger.")

	verify := func() {
		require.Contains(t, buf.String(), `"msg":"foo 42"`, "Unexpected output from Print-family method.")
		buf.Reset()
	}

	std.Print("foo ", 42)
	verify()

	std.Printf("foo %d", 42)
	verify()

	std.Println("foo ", 42)
	verify()
}

func TestStandardLoggerPanic(t *testing.T) {
	std, buf, err := newStd(zap.InfoLevel)
	require.NoError(t, err, "Unexpected error standardizing a Logger.")

	verify := func(f func()) {
		require.Panics(t, f, "Expected calls to Panic methods to panic.")
		require.Contains(t, buf.String(), `"msg":"foo 42"`, "Unexpected output from Panic-family method.")
		buf.Reset()
	}

	verify(func() {
		std.Panic("foo ", 42)
	})

	verify(func() {
		std.Panicf("foo %d", 42)
	})

	verify(func() {
		std.Panicln("foo ", 42)
	})
}

func TestStandardLoggerFatal(t *testing.T) {
	std, buf, err := newStd(zap.InfoLevel)
	require.NoError(t, err, "Unexpected error standardizing a Logger.")

	// Don't actually call os.Exit.
	concrete := std.(*stdLogger)
	concrete.fatal = concrete.write

	verify := func() {
		require.Contains(t, buf.String(), `"msg":"foo 42"`, "Unexpected output from Fatal-family method.")
		buf.Reset()
	}

	std.Fatal("foo ", 42)
	verify()

	std.Fatalf("foo %d", 42)
	verify()

	std.Fatalln("foo ", 42)
	verify()
}
