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

package encoder

import (
	"bytes"
	"io"
	"testing"
	"time"

	"github.com/uber-common/zap/spy"

	"github.com/stretchr/testify/assert"
)

func assertJSON(t *testing.T, expected string, enc *JSONEncoder) {
	assert.Equal(t, expected, string(enc.bytes), "Encoded JSON didn't match expectations.")
}

func withEncoder(f func(*JSONEncoder)) {
	enc := NewJSON()
	enc.AddString("foo", "bar")
	f(enc)
	enc.Free()
}

func TestAddString(t *testing.T) {
	withEncoder(func(enc *JSONEncoder) {
		enc.AddString("baz", "bing")
		assertJSON(t, `"foo":"bar","baz":"bing"`, enc)

		// Keys and values should be escaped.
		enc.truncate()
		enc.AddString(`foo\`, `bar\`)
		assertJSON(t, `"foo\\":"bar\\"`, enc)
	})
}

func TestAddBool(t *testing.T) {
	withEncoder(func(enc *JSONEncoder) {
		enc.AddBool("baz", true)
		assertJSON(t, `"foo":"bar","baz":true`, enc)

		// Keys should be escaped.
		enc.truncate()
		enc.AddBool(`foo\`, false)
		assertJSON(t, `"foo\\":false`, enc)
	})
}

func TestAddInt(t *testing.T) {
	withEncoder(func(enc *JSONEncoder) {
		enc.AddInt("baz", 2)
		assertJSON(t, `"foo":"bar","baz":2`, enc)

		// Keys should be escaped.
		enc.truncate()
		enc.AddInt(`foo\`, 1)
		assertJSON(t, `"foo\\":1`, enc)
	})
}

func TestAddInt64(t *testing.T) {
	withEncoder(func(enc *JSONEncoder) {
		enc.AddInt64("baz", 2)
		assertJSON(t, `"foo":"bar","baz":2`, enc)

		// Keys should be escaped.
		enc.truncate()
		enc.AddInt64(`foo\`, 1)
		assertJSON(t, `"foo\\":1`, enc)
	})
}

func TestAddTime(t *testing.T) {
	withEncoder(func(enc *JSONEncoder) {
		enc.AddTime("ts", time.Unix(0, 100))
		assertJSON(t, `"foo":"bar","ts":100`, enc)

		// Keys should be escaped.
		enc.truncate()
		enc.AddTime(`start\`, time.Unix(0, 0))
		assertJSON(t, `"start\\":0`, enc)
	})
}

func TestAddFloat64(t *testing.T) {
	withEncoder(func(enc *JSONEncoder) {
		enc.AddFloat64("baz", 1e10)
		assertJSON(t, `"foo":"bar","baz":1e+10`, enc)

		// Keys should be escaped.
		enc.truncate()
		enc.AddFloat64(`foo\`, 1.0)
		assertJSON(t, `"foo\\":1`, enc)
	})
}

func TestUnsafeAddBytes(t *testing.T) {
	withEncoder(func(enc *JSONEncoder) {
		enc.UnsafeAddBytes("baz", []byte(`"bing"`))
		assertJSON(t, `"foo":"bar","baz":"bing"`, enc)

		// Keys should be escaped, but values shouldn't.
		enc.truncate()
		enc.UnsafeAddBytes(`foo\`, []byte(`"bar\"`))
		assertJSON(t, `"foo\\":"bar\"`, enc)
	})
}

func TestWriteMessage(t *testing.T) {
	withEncoder(func(enc *JSONEncoder) {
		sink := bytes.NewBuffer(nil)

		// Messages should be escaped.
		err := enc.WriteMessage(sink, "info", `hello\`, time.Unix(0, 0))
		assert.NoError(t, err, "WriteMessage returned an unexpected error.")
		assert.Equal(
			t,
			`{"msg":"hello\\","level":"info","ts":0,"fields":{"foo":"bar"}}`,
			sink.String(),
		)

		// We should be able to re-use the encoder, preserving the accumulated
		// fields.
		sink.Reset()
		err = enc.WriteMessage(sink, "debug", "fake msg", time.Unix(0, 100))
		assert.NoError(t, err, "WriteMessage returned an unexpected error.")
		assert.Equal(
			t,
			`{"msg":"fake msg","level":"debug","ts":100,"fields":{"foo":"bar"}}`,
			sink.String(),
		)
	})
}

func TestWriteMessageFailure(t *testing.T) {
	withEncoder(func(enc *JSONEncoder) {
		tests := []struct {
			sink io.Writer
			msg  string
		}{
			{spy.FailWriter{}, "Expected an error when writing to sink fails."},
			{spy.ShortWriter{}, "Expected an error on partial writes to sink."},
		}
		for _, tt := range tests {
			err := enc.WriteMessage(tt.sink, "info", "hello", time.Unix(0, 0))
			assert.Error(t, err, tt.msg)
		}
	})
}

func TestJSONEscaping(t *testing.T) {
	// Test all the edge cases of JSON escaping directly.
	cases := map[string]string{
		// ASCII.
		`foo`: `foo`,
		// Special-cased characters.
		`"`: `\"`,
		`\`: `\\`,
		`/`: `\/`,
		// Special-cased characters within everyday ASCII.
		`foo"foo`: `foo\"foo`,
		"foo\n":   `foo\n`,
		// Special-cased control characters.
		"\n": `\n`,
		"\r": `\r`,
		"\t": `\t`,
		// \b and \f are special-cased in the JSON spec, but this representation
		// is also conformant.
		"\b": `\u0008`,
		"\f": `\u000c`,
		// The standard lib special-cases angle brackets and ampersands, because
		// it wants to protect users from browser exploits. In a logging
		// context, we shouldn't special-case these characters.
		"<": "<",
		">": ">",
		"&": "&",
		// ASCII bell - not special-cased.
		string(byte(0x07)): `\u0007`,
		// Astral-plane unicode.
		`☃`: `☃`,
		// Decodes to (RuneError, 1)
		"\xed\xa0\x80":    `\ufffd\ufffd\ufffd`,
		"foo\xed\xa0\x80": `foo\ufffd\ufffd\ufffd`,
	}
	enc := NewJSON()
	for input, output := range cases {
		enc.truncate()
		enc.safeAddString(input)
		assertJSON(t, output, enc)
	}
}
