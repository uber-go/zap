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
	"fmt"
	"io"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/uber-go/zap/spywrite"
)

func newTextEncoder(opts ...TextOption) *textEncoder {
	return NewTextEncoder(opts...).(*textEncoder)
}

func withTextEncoder(f func(*textEncoder)) {
	enc := newTextEncoder()
	f(enc)
	enc.Free()
}

func assertTextOutput(t testing.TB, desc string, expected string, f func(Encoder)) {
	withTextEncoder(func(enc *textEncoder) {
		f(enc)
		assert.Equal(t, expected, string(enc.bytes), "Unexpected encoder output after adding a %s.", desc)
	})
	withTextEncoder(func(enc *textEncoder) {
		enc.AddString("foo", "bar")
		f(enc)
		expectedPrefix := "foo=bar"
		if expected != "" {
			// If we expect output, it should be space-separated from the previous
			// field.
			expectedPrefix += " "
		}
		assert.Equal(t, expectedPrefix+expected, string(enc.bytes), "Unexpected encoder output after adding a %s as a second field.", desc)
	})
}

func TestTextEncoderFields(t *testing.T) {
	tests := []struct {
		desc     string
		expected string
		f        func(Encoder)
	}{
		{"string", "k=v", func(e Encoder) { e.AddString("k", "v") }},
		{"string", "k=", func(e Encoder) { e.AddString("k", "") }},
		{"bool", "k=true", func(e Encoder) { e.AddBool("k", true) }},
		{"bool", "k=false", func(e Encoder) { e.AddBool("k", false) }},
		{"int", "k=42", func(e Encoder) { e.AddInt("k", 42) }},
		{"int64", "k=42", func(e Encoder) { e.AddInt64("k", 42) }},
		{"int64", fmt.Sprintf("k=%d", math.MaxInt64), func(e Encoder) { e.AddInt64("k", math.MaxInt64) }},
		{"uint", "k=42", func(e Encoder) { e.AddUint("k", 42) }},
		{"uint64", "k=42", func(e Encoder) { e.AddUint64("k", 42) }},
		{"uint64", fmt.Sprintf("k=%d", uint64(math.MaxUint64)), func(e Encoder) { e.AddUint64("k", math.MaxUint64) }},
		{"uintptr", "k=0xdeadbeef", func(e Encoder) { e.AddUintptr("k", 0xdeadbeef) }},
		{"float64", "k=1", func(e Encoder) { e.AddFloat64("k", 1.0) }},
		{"float64", "k=10000000000", func(e Encoder) { e.AddFloat64("k", 1e10) }},
		{"float64", "k=NaN", func(e Encoder) { e.AddFloat64("k", math.NaN()) }},
		{"float64", "k=+Inf", func(e Encoder) { e.AddFloat64("k", math.Inf(1)) }},
		{"float64", "k=-Inf", func(e Encoder) { e.AddFloat64("k", math.Inf(-1)) }},
		{"marshaler", "k={loggable=yes}", func(e Encoder) {
			assert.NoError(t, e.AddMarshaler("k", loggable{true}), "Unexpected error calling MarshalLog.")
		}},
		{"marshaler", "k={}", func(e Encoder) {
			assert.Error(t, e.AddMarshaler("k", loggable{false}), "Expected an error calling MarshalLog.")
		}},
		{"map[string]string", "k=map[loggable:yes]", func(e Encoder) {
			assert.NoError(t, e.AddObject("k", map[string]string{"loggable": "yes"}), "Unexpected error serializing a map.")
		}},
		{"arbitrary object", "k={Name:jane}", func(e Encoder) {
			assert.NoError(t, e.AddObject("k", struct{ Name string }{"jane"}), "Unexpected error serializing a struct.")
		}},
	}

	for _, tt := range tests {
		assertTextOutput(t, tt.desc, tt.expected, tt.f)
	}
}

func TestTextWriteEntry(t *testing.T) {
	entry := &Entry{Level: InfoLevel, Message: "Something happened.", Time: epoch}
	tests := []struct {
		enc      Encoder
		expected string
		name     string
	}{
		{
			enc:      NewTextEncoder(),
			expected: "[I] 1970-01-01T00:00:00Z Something happened.",
			name:     "RFC822",
		},
		{
			enc:      NewTextEncoder(TextTimeFormat(time.RFC822)),
			expected: "[I] 01 Jan 70 00:00 UTC Something happened.",
			name:     "RFC822",
		},
		{
			enc:      NewTextEncoder(TextTimeFormat("")),
			expected: "[I] Something happened.",
			name:     "empty layout",
		},
		{
			enc:      NewTextEncoder(TextNoTime()),
			expected: "[I] Something happened.",
			name:     "NoTime",
		},
	}

	sink := &testBuffer{}
	for _, tt := range tests {
		assert.NoError(
			t,
			tt.enc.WriteEntry(sink, entry.Message, entry.Level, entry.Time),
			"Unexpected failure writing entry with text time formatter %s.", tt.name,
		)
		assert.Equal(t, tt.expected, sink.Stripped(), "Unexpected output from text time formatter %s.", tt.name)
		sink.Reset()
	}
}

func TestTextWriteEntryLevels(t *testing.T) {
	tests := []struct {
		level    Level
		expected string
	}{
		{DebugLevel, "D"},
		{InfoLevel, "I"},
		{WarnLevel, "W"},
		{ErrorLevel, "E"},
		{PanicLevel, "P"},
		{FatalLevel, "F"},
		{Level(42), "42"},
	}

	sink := &testBuffer{}
	enc := NewTextEncoder(TextNoTime())
	for _, tt := range tests {
		assert.NoError(
			t,
			enc.WriteEntry(sink, "Fake message.", tt.level, epoch),
			"Unexpected failure writing entry with level %s.", tt.level,
		)
		expected := fmt.Sprintf("[%s] Fake message.", tt.expected)
		assert.Equal(t, expected, sink.Stripped(), "Unexpected text output for level %s.", tt.level)
		sink.Reset()
	}
}

func TestTextClone(t *testing.T) {
	parent := &textEncoder{bytes: make([]byte, 0, 128)}
	clone := parent.Clone()

	// Adding to the parent shouldn't affect the clone, and vice versa.
	parent.AddString("foo", "bar")
	clone.AddString("baz", "bing")

	assert.Equal(t, "foo=bar", string(parent.bytes), "Unexpected serialized fields in parent encoder.")
	assert.Equal(t, "baz=bing", string(clone.(*textEncoder).bytes), "Unexpected serialized fields in cloned encoder.")
}

func TestTextWriteEntryFailure(t *testing.T) {
	withTextEncoder(func(enc *textEncoder) {
		tests := []struct {
			sink io.Writer
			msg  string
		}{
			{nil, "Expected an error when writing to a nil sink."},
			{spywrite.FailWriter{}, "Expected an error when writing to sink fails."},
			{spywrite.ShortWriter{}, "Expected an error on partial writes to sink."},
		}
		for _, tt := range tests {
			err := enc.WriteEntry(tt.sink, "hello", InfoLevel, time.Unix(0, 0))
			assert.Error(t, err, tt.msg)
		}
	})
}

func TestTextTimeOptions(t *testing.T) {
	epoch := time.Date(1970, time.January, 1, 0, 0, 0, 0, time.UTC)
	entry := &Entry{Level: InfoLevel, Message: "Something happened.", Time: epoch}

	enc := NewTextEncoder()

	sink := &testBuffer{}
	enc.AddString("foo", "bar")
	err := enc.WriteEntry(sink, entry.Message, entry.Level, entry.Time)
	assert.NoError(t, err, "WriteEntry returned an unexpected error.")
	assert.Equal(
		t,
		"[I] 1970-01-01T00:00:00Z Something happened. foo=bar",
		sink.Stripped(),
	)
}
