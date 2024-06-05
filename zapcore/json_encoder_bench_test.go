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
	"encoding/json"
	"fmt"
	"testing"
	"time"

	//revive:disable:dot-imports
	. "go.uber.org/zap/zapcore"
)

func BenchmarkJSONLogMarshalerFunc(b *testing.B) {
	for i := 0; i < b.N; i++ {
		enc := NewJSONEncoder(testEncoderConfig())
		err := enc.AddObject("nested", ObjectMarshalerFunc(func(enc ObjectEncoder) error {
			enc.AddInt64("i", int64(i))
			return nil
		}))
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkZapJSONFloat32AndComplex64(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			enc := NewJSONEncoder(testEncoderConfig())
			enc.AddFloat32("float32", 3.14)
			enc.AddComplex64("complex64", 2.71+3.14i)
		}
	})
}

const _sliceSize = 5000

type StringSlice []string

func (s StringSlice) MarshalLogArray(encoder ArrayEncoder) error {
	for _, str := range s {
		encoder.AppendString(str)
	}
	return nil
}

func generateStringSlice(n int) StringSlice {
	output := make(StringSlice, 0, n)
	for i := 0; i < n; i++ {
		output = append(output, fmt.Sprint("00000000-0000-0000-0000-0000000000", i))
	}
	return output
}

func BenchmarkZapJSON(b *testing.B) {
	additional := generateStringSlice(_sliceSize)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			enc := NewJSONEncoder(testEncoderConfig())
			enc.AddString("str", "foo")
			enc.AddInt64("int64-1", 1)
			enc.AddInt64("int64-2", 2)
			enc.AddFloat64("float64", 1.0)
			enc.AddString("string1", "\n")
			enc.AddString("string2", "💩")
			enc.AddString("string3", "🤔")
			enc.AddString("string4", "🙊")
			enc.AddBool("bool", true)
			_ = enc.AddArray("test", additional)
			buf, _ := enc.EncodeEntry(Entry{
				Message: "fake",
				Level:   DebugLevel,
			}, nil)
			buf.Free()
		}
	})
}

func BenchmarkStandardJSON(b *testing.B) {
	record := struct {
		Level      string                 `json:"level"`
		Message    string                 `json:"msg"`
		Time       time.Time              `json:"ts"`
		Fields     map[string]interface{} `json:"fields"`
		Additional StringSlice
	}{
		Level:   "debug",
		Message: "fake",
		Time:    time.Unix(0, 0),
		Fields: map[string]interface{}{
			"str":     "foo",
			"int64-1": int64(1),
			"int64-2": int64(1),
			"float64": float64(1.0),
			"string1": "\n",
			"string2": "💩",
			"string3": "🤔",
			"string4": "🙊",
			"bool":    true,
		},
		Additional: generateStringSlice(_sliceSize),
	}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if _, err := json.Marshal(record); err != nil {
				b.Fatal(err)
			}
		}
	})
}
