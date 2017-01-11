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

package zapcore

import (
	"errors"
	"fmt"
	"io"
	"math"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"go.uber.org/zap/testutils"
)

var epoch = time.Date(1970, time.January, 1, 0, 0, 0, 0, time.UTC)

func newJSONEncoder(cfg JSONConfig) *jsonEncoder {
	return NewJSONEncoder(cfg).(*jsonEncoder)
}

func assertJSON(t *testing.T, expected string, enc *jsonEncoder) {
	assert.Equal(t, expected, string(enc.bytes), "Encoded JSON didn't match expectations.")
}

func withJSONEncoder(f func(*jsonEncoder)) {
	enc := newJSONEncoder(testJSONConfig())
	f(enc)
	enc.free()
}

type noJSON struct{}

func (nj noJSON) MarshalJSON() ([]byte, error) {
	return nil, errors.New("no")
}

type loggable struct{ bool }

func (l loggable) MarshalLogObject(enc ObjectEncoder) error {
	if !l.bool {
		return errors.New("can't marshal")
	}
	enc.AddString("loggable", "yes")
	return nil
}

func assertOutput(t testing.TB, desc string, expected string, f func(Encoder)) {
	withJSONEncoder(func(enc *jsonEncoder) {
		f(enc)
		assert.Equal(t, expected, string(enc.bytes), "Unexpected encoder output after adding a %s.", desc)
	})
	withJSONEncoder(func(enc *jsonEncoder) {
		enc.AddString("foo", "bar")
		f(enc)
		expectedPrefix := `"foo":"bar"`
		if expected != "" {
			// If we expect output, it should be comma-separated from the previous
			// field.
			expectedPrefix += ","
		}
		assert.Equal(t, expectedPrefix+expected, string(enc.bytes), "Unexpected encoder output after adding a %s as a second field.", desc)
	})
}

func TestJSONEncoderFields(t *testing.T) {
	tests := []struct {
		desc     string
		expected string
		f        func(Encoder)
	}{
		{"string", `"k":"v"`, func(e Encoder) { e.AddString("k", "v") }},
		{"string", `"k":""`, func(e Encoder) { e.AddString("k", "") }},
		{"string", `"k\\":"v\\"`, func(e Encoder) { e.AddString(`k\`, `v\`) }},
		{"bool", `"k":true`, func(e Encoder) { e.AddBool("k", true) }},
		{"bool", `"k":false`, func(e Encoder) { e.AddBool("k", false) }},
		{"bool", `"k\\":true`, func(e Encoder) { e.AddBool(`k\`, true) }},
		{"int64", `"k":42`, func(e Encoder) { e.AddInt64("k", 42) }},
		{"int64", `"k\\":42`, func(e Encoder) { e.AddInt64(`k\`, 42) }},
		{"int64", fmt.Sprintf(`"k":%d`, math.MaxInt64), func(e Encoder) { e.AddInt64("k", math.MaxInt64) }},
		{"int64", fmt.Sprintf(`"k":%d`, math.MinInt64), func(e Encoder) { e.AddInt64("k", math.MinInt64) }},
		{"int64", fmt.Sprintf(`"k\\":%d`, math.MaxInt64), func(e Encoder) { e.AddInt64(`k\`, math.MaxInt64) }},
		{"uint64", `"k":42`, func(e Encoder) { e.AddUint64("k", 42) }},
		{"uint64", `"k\\":42`, func(e Encoder) { e.AddUint64(`k\`, 42) }},
		{"uint64", fmt.Sprintf(`"k":%d`, uint64(math.MaxUint64)), func(e Encoder) { e.AddUint64("k", math.MaxUint64) }},
		{"uint64", fmt.Sprintf(`"k\\":%d`, uint64(math.MaxUint64)), func(e Encoder) { e.AddUint64(`k\`, math.MaxUint64) }},
		{"float64", `"k":1`, func(e Encoder) { e.AddFloat64("k", 1.0) }},
		{"float64", `"k\\":1`, func(e Encoder) { e.AddFloat64(`k\`, 1.0) }},
		{"float64", `"k":10000000000`, func(e Encoder) { e.AddFloat64("k", 1e10) }},
		{"float64", `"k":"NaN"`, func(e Encoder) { e.AddFloat64("k", math.NaN()) }},
		{"float64", `"k":"+Inf"`, func(e Encoder) { e.AddFloat64("k", math.Inf(1)) }},
		{"float64", `"k":"-Inf"`, func(e Encoder) { e.AddFloat64("k", math.Inf(-1)) }},
		{"ObjectMarshaler", `"k":{"loggable":"yes"}`, func(e Encoder) {
			assert.NoError(t, e.AddObject("k", loggable{true}), "Unexpected error calling MarshalObject.")
		}},
		{"ObjectMarshaler", `"k\\":{"loggable":"yes"}`, func(e Encoder) {
			assert.NoError(t, e.AddObject(`k\`, loggable{true}), "Unexpected error calling MarshalObject.")
		}},
		{"ObjectMarshaler", `"k":{}`, func(e Encoder) {
			assert.Error(t, e.AddObject("k", loggable{false}), "Expected an error calling MarshalObject.")
		}},
		{"arbitrary object", `"k":{"loggable":"yes"}`, func(e Encoder) {
			assert.NoError(t, e.AddReflected("k", map[string]string{"loggable": "yes"}), "Unexpected error JSON-serializing a map.")
		}},
		{"arbitrary object", `"k\\":{"loggable":"yes"}`, func(e Encoder) {
			assert.NoError(t, e.AddReflected(`k\`, map[string]string{"loggable": "yes"}), "Unexpected error JSON-serializing a map.")
		}},
		{"arbitrary object", "", func(e Encoder) {
			assert.Error(t, e.AddReflected("k", noJSON{}), "Unexpected success JSON-serializing a noJSON.")
		}},
	}

	for _, tt := range tests {
		assertOutput(t, tt.desc, tt.expected, tt.f)
	}
}

func TestJSONWriteEntry(t *testing.T) {
	epoch := time.Unix(0, 0)
	withJSONEncoder(func(enc *jsonEncoder) {
		// If we somehow have a nil sink, we should return an error instead of
		// panicing.
		assert.Equal(t, errNilSink, enc.WriteEntry(nil, Entry{
			Level:   InfoLevel,
			Message: `hello\`,
			Time:    epoch,
		}, nil), "Expected an error writing to a nil sink.")

		// Messages should be escaped.
		sink := &testutils.Buffer{}
		enc.AddString("foo", "bar")
		err := enc.WriteEntry(sink, Entry{
			Level:   InfoLevel,
			Message: `hello\`,
			Time:    epoch,
		}, nil)
		assert.NoError(t, err, "WriteEntry returned an unexpected error.")
		assert.Equal(
			t,
			`{"level":"info","ts":0,"msg":"hello\\","foo":"bar"}`,
			sink.Stripped(),
		)

		// We should be able to re-use the encoder, preserving the accumulated
		// fields.
		sink.Reset()
		err = enc.WriteEntry(sink, Entry{
			Level:   InfoLevel,
			Message: `hello`,
			Time:    time.Unix(0, 100*int64(time.Millisecond)),
		}, nil)
		assert.NoError(t, err, "WriteEntry returned an unexpected error.")
		assert.Equal(
			t,
			`{"level":"info","ts":100,"msg":"hello","foo":"bar"}`,
			sink.Stripped(),
		)

		// Stacktraces are included.
		sink.Reset()
		err = enc.WriteEntry(sink, Entry{
			Level:   InfoLevel,
			Message: `hello`,
			Time:    time.Unix(0, 100*int64(time.Millisecond)),
			Stack:   "trace",
		}, nil)
		assert.NoError(t, err, "WriteEntry returned an unexpected error.")
		assert.Equal(
			t,
			`{"level":"info","ts":100,"msg":"hello","foo":"bar","stacktrace":"trace"}`,
			sink.Stripped(),
		)

		// Caller is included.
		sink.Reset()
		err = enc.WriteEntry(sink, Entry{
			Level:   InfoLevel,
			Message: `hello`,
			Time:    time.Unix(0, 100*int64(time.Millisecond)),
			Caller:  MakeEntryCaller(runtime.Caller(0)),
		}, nil)
		assert.NoError(t, err, "WriteEntry returned an unexpected error.")
		assert.Regexp(
			t,
			`{"level":"info","ts":100,"caller":"/.*zap/zapcore/json_encoder_test.go:\d+","msg":"hello","foo":"bar"}`,
			sink.Stripped(),
		)
	})
}

func TestJSONClone(t *testing.T) {
	// The parent encoder is created with plenty of excess capacity.
	cfg := testJSONConfig()
	parent := &jsonEncoder{
		JSONConfig: &cfg,
		bytes:      make([]byte, 0, 128),
	}
	clone := parent.Clone()

	// Adding to the parent shouldn't affect the clone, and vice versa.
	parent.AddString("foo", "bar")
	clone.AddString("baz", "bing")

	assertJSON(t, `"foo":"bar"`, parent)
	assertJSON(t, `"baz":"bing"`, clone.(*jsonEncoder))
}

func TestJSONWriteEntryFailure(t *testing.T) {
	withJSONEncoder(func(enc *jsonEncoder) {
		tests := []struct {
			sink io.Writer
			msg  string
		}{
			{&testutils.FailWriter{}, "Expected an error when writing to sink fails."},
			{&testutils.ShortWriter{}, "Expected an error on partial writes to sink."},
		}
		for _, tt := range tests {
			err := enc.WriteEntry(tt.sink, Entry{
				Message: "hello",
				Level:   InfoLevel,
				Time:    time.Unix(0, 0),
			}, nil)
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
		// Special-cased characters within everyday ASCII.
		`foo"foo`: `foo\"foo`,
		"foo\n":   `foo\n`,
		// Special-cased control characters.
		"\n": `\n`,
		"\r": `\r`,
		"\t": `\t`,
		// \b and \f are sometimes backslash-escaped, but this representation is also
		// conformant.
		"\b": `\u0008`,
		"\f": `\u000c`,
		// The standard lib special-cases angle brackets and ampersands by default,
		// because it wants to protect users from browser exploits. In a logging
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
	withJSONEncoder(func(enc *jsonEncoder) {
		for input, output := range cases {
			enc.truncate()
			enc.safeAddString(input)
			assertJSON(t, output, enc)
		}
	})
}
