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
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"reflect"
	"testing"
	"testing/quick"

	"go.uber.org/zap/internal/multierror"

	"github.com/stretchr/testify/assert"
)

func TestJSONClone(t *testing.T) {
	// The parent encoder is created with plenty of excess capacity.
	parent := &jsonEncoder{
		bytes: make([]byte, 0, 128),
	}
	clone := parent.Clone()

	// Adding to the parent shouldn't affect the clone, and vice versa.
	parent.AddString("foo", "bar")
	clone.AddString("baz", "bing")

	assertJSON(t, `"foo":"bar"`, parent)
	assertJSON(t, `"baz":"bing"`, clone.(*jsonEncoder))
}

func TestJSONEscaping(t *testing.T) {
	enc := &jsonEncoder{}
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
	for input, output := range cases {
		enc.truncate()
		enc.safeAddString(input)
		assertJSON(t, output, enc)
	}
}

func TestJSONEncoderObjectFields(t *testing.T) {
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
			assert.NoError(t, e.AddObject("k", loggable{true}), "Unexpected error calling MarshalLogObject.")
		}},
		{"ObjectMarshaler", `"k\\":{"loggable":"yes"}`, func(e Encoder) {
			assert.NoError(t, e.AddObject(`k\`, loggable{true}), "Unexpected error calling MarshalLogObject.")
		}},
		{"ObjectMarshaler", `"k":{}`, func(e Encoder) {
			assert.Error(t, e.AddObject("k", loggable{false}), "Expected an error calling MarshalLogObject.")
		}},
		{
			"ObjectMarshaler(ArrayMarshaler(ObjectMarshaler))",
			`"turducken":{"ducks":[{"in":"chicken"},{"in":"chicken"}]}`,
			func(e Encoder) {
				assert.NoError(
					t,
					e.AddObject("turducken", turducken{}),
					"Unexpected error calling MarshalLogObject with nested ObjectMarshalers and ArrayMarshalers.",
				)
			},
		},
		{
			"ArrayMarshaler(ObjectMarshaler(ArrayMarshaler(ObjectMarshaler)))",
			`"turduckens":[{"ducks":[{"in":"chicken"},{"in":"chicken"}]},{"ducks":[{"in":"chicken"},{"in":"chicken"}]}]`,
			func(e Encoder) {
				assert.NoError(
					t,
					e.AddArray("turduckens", turduckens(2)),
					"Unexpected error calling MarshalLogObject with nested ObjectMarshalers and ArrayMarshalers.",
				)
			},
		},
		{"ArrayMarshaler", `"k\\":[true]`, func(e Encoder) {
			assert.NoError(t, e.AddArray(`k\`, loggable{true}), "Unexpected error calling MarshalLogArray.")
		}},
		{"ArrayMarshaler", `"k":[]`, func(e Encoder) {
			assert.Error(t, e.AddArray("k", loggable{false}), "Expected an error calling MarshalLogArray.")
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

func TestJSONEncoderArrayTypes(t *testing.T) {
	tests := []struct {
		desc        string
		f           func(ArrayEncoder) error
		expected    string
		shouldError bool
	}{
		// arrays of ObjectMarshalers are covered by the turducken test above.
		{
			"arrays of arrays",
			func(arr ArrayEncoder) error {
				arr.AppendArray(ArrayMarshalerFunc(func(enc ArrayEncoder) error {
					enc.AppendBool(true)
					return nil
				}))
				arr.AppendArray(ArrayMarshalerFunc(func(enc ArrayEncoder) error {
					enc.AppendBool(true)
					return nil
				}))
				return nil
			},
			`[[true],[true]]`,
			false,
		},
		{
			"bools",
			func(arr ArrayEncoder) error {
				arr.AppendBool(true)
				arr.AppendBool(false)
				return nil
			},
			`[true,false]`,
			false,
		},
	}

	for _, tt := range tests {
		f := func(enc Encoder) error {
			return enc.AddArray("array", ArrayMarshalerFunc(tt.f))
		}
		assertOutput(t, tt.desc, `"array":`+tt.expected, func(enc Encoder) {
			err := f(enc)
			if tt.shouldError {
				assert.Error(t, err, "Expected an error adding array to JSON encoder.")
			} else {
				assert.NoError(t, err, "Unexpected error adding array to JSON encoder.")
			}
		})
	}
}

func assertJSON(t *testing.T, expected string, enc *jsonEncoder) {
	assert.Equal(t, expected, string(enc.bytes), "Encoded JSON didn't match expectations.")
}

func assertOutput(t testing.TB, desc string, expected string, f func(Encoder)) {
	enc := &jsonEncoder{}
	f(enc)
	assert.Equal(t, expected, string(enc.bytes), "Unexpected encoder output after adding a %s.", desc)

	enc = &jsonEncoder{}
	enc.AddString("foo", "bar")
	f(enc)
	expectedPrefix := `"foo":"bar"`
	if expected != "" {
		// If we expect output, it should be comma-separated from the previous
		// field.
		expectedPrefix += ","
	}
	assert.Equal(t, expectedPrefix+expected, string(enc.bytes), "Unexpected encoder output after adding a %s as a second field.", desc)
}

// Nested Array- and ObjectMarshalers.
type turducken struct{}

func (t turducken) MarshalLogObject(enc ObjectEncoder) error {
	return enc.AddArray("ducks", ArrayMarshalerFunc(func(arr ArrayEncoder) error {
		for i := 0; i < 2; i++ {
			arr.AppendObject(ObjectMarshalerFunc(func(inner ObjectEncoder) error {
				inner.AddString("in", "chicken")
				return nil
			}))
		}
		return nil
	}))
}

type turduckens int

func (t turduckens) MarshalLogArray(enc ArrayEncoder) error {
	var errs multierror.Error
	tur := turducken{}
	for i := 0; i < int(t); i++ {
		errs = errs.Append(enc.AppendObject(tur))
	}
	return errs.AsError()
}

type loggable struct{ bool }

func (l loggable) MarshalLogObject(enc ObjectEncoder) error {
	if !l.bool {
		return errors.New("can't marshal")
	}
	enc.AddString("loggable", "yes")
	return nil
}

func (l loggable) MarshalLogArray(enc ArrayEncoder) error {
	if !l.bool {
		return errors.New("can't marshal")
	}
	enc.AppendBool(true)
	return nil
}

type noJSON struct{}

func (nj noJSON) MarshalJSON() ([]byte, error) {
	return nil, errors.New("no")
}

func zapEncodeString(s string) []byte {
	enc := &jsonEncoder{}
	// Escape and quote a string using our encoder.
	var ret []byte
	enc.safeAddString(s)
	ret = make([]byte, 0, len(enc.bytes)+2)
	ret = append(ret, '"')
	ret = append(ret, enc.bytes...)
	ret = append(ret, '"')
	return ret
}

func roundTripsCorrectly(original string) bool {
	// Encode using our encoder, decode using the standard library, and assert
	// that we haven't lost any information.
	encoded := zapEncodeString(original)

	var decoded string
	err := json.Unmarshal(encoded, &decoded)
	if err != nil {
		return false
	}
	return original == decoded
}

type ASCII string

func (s ASCII) Generate(r *rand.Rand, size int) reflect.Value {
	bs := make([]byte, size)
	for i := range bs {
		bs[i] = byte(r.Intn(128))
	}
	a := ASCII(bs)
	return reflect.ValueOf(a)
}

func asciiRoundTripsCorrectly(s ASCII) bool {
	return roundTripsCorrectly(string(s))
}

func TestJSONQuick(t *testing.T) {
	// Test the full range of UTF-8 strings.
	err := quick.Check(roundTripsCorrectly, &quick.Config{MaxCountScale: 100.0})
	if err != nil {
		t.Error(err.Error())
	}

	// Focus on ASCII strings.
	err = quick.Check(asciiRoundTripsCorrectly, &quick.Config{MaxCountScale: 100.0})
	if err != nil {
		t.Error(err.Error())
	}
}
