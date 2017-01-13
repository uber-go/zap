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
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	. "go.uber.org/zap/zapcore"
)

var epoch = time.Date(1970, time.January, 1, 0, 0, 0, 0, time.UTC)

func withJSONEncoder(f func(Encoder)) {
	f(NewJSONEncoder(testJSONConfig()))
}

func TestJSONEncodeEntry(t *testing.T) {
	epoch := time.Unix(0, 0)
	withJSONEncoder(func(enc Encoder) {
		// Messages should be escaped.
		enc.AddString("foo", "bar")
		buf, err := enc.EncodeEntry(Entry{
			Level:   InfoLevel,
			Message: `hello\`,
			Time:    epoch,
		}, nil)
		assert.NoError(t, err, "EncodeEntry returned an unexpected error.")
		assert.Equal(
			t,
			`{"level":"info","ts":0,"msg":"hello\\","foo":"bar"}`+"\n",
			string(buf),
		)

		// We should be able to re-use the encoder, preserving the accumulated
		// fields.
		buf, err = enc.EncodeEntry(Entry{
			Level:   InfoLevel,
			Message: `hello`,
			Time:    time.Unix(0, 100*int64(time.Millisecond)),
		}, nil)
		assert.NoError(t, err, "EncodeEntry returned an unexpected error.")
		assert.Equal(
			t,
			`{"level":"info","ts":100,"msg":"hello","foo":"bar"}`+"\n",
			string(buf),
		)

		// Stacktraces are included.
		buf, err = enc.EncodeEntry(Entry{
			Level:   InfoLevel,
			Message: `hello`,
			Time:    time.Unix(0, 100*int64(time.Millisecond)),
			Stack:   "trace",
		}, nil)
		assert.NoError(t, err, "EncodeEntry returned an unexpected error.")
		assert.Equal(
			t,
			`{"level":"info","ts":100,"msg":"hello","foo":"bar","stacktrace":"trace"}`+"\n",
			string(buf),
		)

		// Caller is included.
		buf, err = enc.EncodeEntry(Entry{
			Level:   InfoLevel,
			Message: `hello`,
			Time:    time.Unix(0, 100*int64(time.Millisecond)),
			Caller:  MakeEntryCaller(runtime.Caller(0)),
		}, nil)
		assert.NoError(t, err, "EncodeEntry returned an unexpected error.")
		assert.Regexp(
			t,
			`{"level":"info","ts":100,"caller":"/.*zap/zapcore/json_encoder_test.go:\d+","msg":"hello","foo":"bar"}`+"\n",
			string(buf),
		)
	})
}
