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
	"bytes"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNullEncoderFields(t *testing.T) {
	ne := NullEncoder()
	for _, tt := range []struct {
		desc string
		f    func(Encoder)
	}{
		{"string", func(e Encoder) { e.AddString("k", "v") }},
		{"bool", func(e Encoder) { e.AddBool("k", true) }},
		{"bool", func(e Encoder) { e.AddBool("k", false) }},
		{"int", func(e Encoder) { e.AddInt("k", 42) }},
		{"int64", func(e Encoder) { e.AddInt64("k", math.MaxInt64) }},
		{"int64", func(e Encoder) { e.AddInt64("k", math.MinInt64) }},
		{"uint", func(e Encoder) { e.AddUint("k", 42) }},
		{"uint64", func(e Encoder) { e.AddUint64("k", math.MaxUint64) }},
		{"uintptr", func(e Encoder) { e.AddUintptr("k", uintptr(math.MaxUint64)) }},
		{"float64", func(e Encoder) { e.AddFloat64("k", 1.0) }},
		{"ObjectMarshaler", func(e Encoder) {
			assert.NoError(t, e.AddObject("k", loggable{true}), "Unexpected error calling MarshalLogObject.")
		}},
		{"arbitrary object", func(e Encoder) {
			assert.NoError(t, e.AddReflected("k", map[string]string{"": ""}), "Unexpected error.")
		}},
	} {
		assert.NotPanics(t, func() { tt.f(ne) }, tt.desc)
	}
}

func TestNullWriteEntry(t *testing.T) {
	enc := NullEncoder()
	assert.Equal(t, errNilSink, enc.WriteEntry(nil, Entry{
		Level:   InfoLevel,
		Message: `ohai`,
		Time:    time.Unix(0, 0),
	}, nil), "Expected an error writing to a nil sink.")

	// Messages should be thrown away.
	sink := &bytes.Buffer{}
	enc.AddString("foo", "bar")
	assert.Len(t, sink.Bytes(), 0)
	err := enc.WriteEntry(sink, Entry{
		Level:   InfoLevel,
		Message: `ohai`,
		Time:    time.Unix(0, 0),
	}, nil)
	assert.NoError(t, err, "WriteEntry returned an unexpected error.")
	assert.Len(
		t,
		sink.Bytes(),
		0,
		"WriteEntry actually wrote something",
	)
}
