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
	"errors"
	"io"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/uber-go/zap/spywrite"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func assertJSON(t *testing.T, expected string, enc *jsonEncoder) {
	assert.Equal(t, expected, string(enc.bytes), "Encoded JSON didn't match expectations.")
}

func withJSONEncoder(f func(*jsonEncoder)) {
	enc := newJSONEncoder()
	enc.AddString("foo", "bar")
	f(enc)
	enc.Free()
}

type noJSON struct{}

func (nj noJSON) MarshalJSON() ([]byte, error) {
	return nil, errors.New("no")
}

func TestJSONAddString(t *testing.T) {
	withJSONEncoder(func(enc *jsonEncoder) {
		enc.AddString("baz", "bing")
		assertJSON(t, `"foo":"bar","baz":"bing"`, enc)

		// Keys and values should be escaped.
		enc.truncate()
		enc.AddString(`foo\`, `bar\`)
		assertJSON(t, `"foo\\":"bar\\"`, enc)
	})
}

func TestJSONAddBool(t *testing.T) {
	withJSONEncoder(func(enc *jsonEncoder) {
		enc.AddBool("baz", true)
		assertJSON(t, `"foo":"bar","baz":true`, enc)

		// Keys should be escaped.
		enc.truncate()
		enc.AddBool(`foo\`, false)
		assertJSON(t, `"foo\\":false`, enc)
	})
}

func TestJSONAddInt(t *testing.T) {
	withJSONEncoder(func(enc *jsonEncoder) {
		enc.AddInt("baz", 2)
		assertJSON(t, `"foo":"bar","baz":2`, enc)

		// Keys should be escaped.
		enc.truncate()
		enc.AddInt(`foo\`, 1)
		assertJSON(t, `"foo\\":1`, enc)
	})
}

func TestJSONAddInt64(t *testing.T) {
	withJSONEncoder(func(enc *jsonEncoder) {
		enc.AddInt64("baz", 2)
		assertJSON(t, `"foo":"bar","baz":2`, enc)

		// Keys should be escaped.
		enc.truncate()
		enc.AddInt64(`foo\`, 1)
		assertJSON(t, `"foo\\":1`, enc)
	})
}

func TestJSONAddFloat64(t *testing.T) {
	withJSONEncoder(func(enc *jsonEncoder) {
		enc.AddFloat64("baz", 1e10)
		assertJSON(t, `"foo":"bar","baz":1e+10`, enc)

		// Keys should be escaped.
		enc.truncate()
		enc.AddFloat64(`foo\`, 1.0)
		assertJSON(t, `"foo\\":1`, enc)

		// Test floats that can't be represented in JSON.
		enc.truncate()
		enc.AddFloat64(`foo`, math.NaN())
		assertJSON(t, `"foo":"NaN"`, enc)

		enc.truncate()
		enc.AddFloat64(`foo`, math.Inf(1))
		assertJSON(t, `"foo":"+Inf"`, enc)

		enc.truncate()
		enc.AddFloat64(`foo`, math.Inf(-1))
		assertJSON(t, `"foo":"-Inf"`, enc)
	})
}

func TestJSONWriteMessage(t *testing.T) {
	withJSONEncoder(func(enc *jsonEncoder) {
		sink := bytes.NewBuffer(nil)

		// Messages should be escaped.
		err := enc.WriteMessage(sink, "info", `hello\`, time.Unix(0, 0))
		assert.NoError(t, err, "WriteMessage returned an unexpected error.")
		assert.Equal(t,
			`{"msg":"hello\\","level":"info","ts":0,"fields":{"foo":"bar"}}`,
			strings.TrimRight(sink.String(), "\n"),
		)

		// We should be able to re-use the encoder, preserving the accumulated
		// fields.
		sink.Reset()
		err = enc.WriteMessage(sink, "debug", "fake msg", time.Unix(0, 100))
		assert.NoError(t, err, "WriteMessage returned an unexpected error.")
		assert.Equal(t,
			`{"msg":"fake msg","level":"debug","ts":100,"fields":{"foo":"bar"}}`,
			strings.TrimRight(sink.String(), "\n"),
		)
	})
}

type loggable struct{}

func (l loggable) MarshalLog(kv KeyValue) error {
	kv.AddString("loggable", "yes")
	return nil
}

func TestJSONAddMarshaler(t *testing.T) {
	withJSONEncoder(func(enc *jsonEncoder) {
		err := enc.AddMarshaler("nested", loggable{})
		require.NoError(t, err, "Unexpected error using AddMarshaler.")
		assertJSON(t, `"foo":"bar","nested":{"loggable":"yes"}`, enc)
	})
}

func TestJSONAddObject(t *testing.T) {
	withJSONEncoder(func(enc *jsonEncoder) {
		enc.AddObject("nested", map[string]string{"loggable": "yes"})
		assertJSON(t, `"foo":"bar","nested":{"loggable":"yes"}`, enc)
	})

	// Serialization errors are handled by the field.
	withJSONEncoder(func(enc *jsonEncoder) {
		require.Error(t, enc.AddObject("nested", noJSON{}), "Unexpected success encoding non-JSON-serializable object.")
		assertJSON(t, `"foo":"bar"`, enc)
	})
}

func TestJSONClone(t *testing.T) {
	// The parent encoder is created with plenty of excess capacity.
	parent := &jsonEncoder{bytes: make([]byte, 0, 128)}
	clone := parent.Clone()

	// Adding to the parent shouldn't affect the clone, and vice versa.
	parent.AddString("foo", "bar")
	clone.AddString("baz", "bing")

	assertJSON(t, `"foo":"bar"`, parent)
	assertJSON(t, `"baz":"bing"`, clone.(*jsonEncoder))
}

func TestJSONWriteMessageFailure(t *testing.T) {
	withJSONEncoder(func(enc *jsonEncoder) {
		tests := []struct {
			sink io.Writer
			msg  string
		}{
			{spywrite.FailWriter{}, "Expected an error when writing to sink fails."},
			{spywrite.ShortWriter{}, "Expected an error on partial writes to sink."},
		}
		for _, tt := range tests {
			err := enc.WriteMessage(tt.sink, "info", "hello", time.Unix(0, 0))
			assert.Error(t, err, tt.msg)
		}
	})
}

func TestJSONJSONEscaping(t *testing.T) {
	// Test all the edge cases of JSON escaping directly.
	cases := map[string]string{
		// ASCII.
		`foo`: `foo`,
		// Special-cased characters.
		`"`: `\"`,
		`\`: `\\`,
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
	enc := newJSONEncoder()
	for input, output := range cases {
		enc.truncate()
		enc.safeAddString(input)
		assertJSON(t, output, enc)
	}
}
