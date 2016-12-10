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
	"encoding/json"
	"io/ioutil"
	"testing"
	"time"
)

type logRecord struct {
	Level   string                 `json:"level"`
	Message string                 `json:"msg"`
	Time    time.Time              `json:"ts"`
	Fields  map[string]interface{} `json:"fields"`
}

func BenchmarkJSONLogMarshalerFunc(b *testing.B) {
	for i := 0; i < b.N; i++ {
		enc := NewJSONEncoder()
		enc.AddMarshaler("nested", LogMarshalerFunc(func(kv KeyValue) error {
			kv.AddInt("i", i)
			return nil
		}))
		enc.Free()
	}
}

func BenchmarkZapJSONInts(b *testing.B) {
	ts := time.Unix(0, 0)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			enc := NewJSONEncoder()
			enc.AddInts("ints", []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})
			enc.WriteEntry(ioutil.Discard, "fake", DebugLevel, ts)
			enc.Free()
		}
	})
}

func BenchmarkZapJSONStrings(b *testing.B) {
	ts := time.Unix(0, 0)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			enc := NewJSONEncoder()
			enc.AddStrings("strings", []string{"bar 1", "bar 2", "bar 3", "bar 4", "bar 5", "bar 6", "bar 7", "bar 8", "bar 9", "bar 10"})
			enc.WriteEntry(ioutil.Discard, "fake", DebugLevel, ts)
			enc.Free()
		}
	})
}

func BenchmarkZapJSON(b *testing.B) {
	ts := time.Unix(0, 0)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			enc := NewJSONEncoder()
			enc.AddString("str", "foo")
			enc.AddInt("int", 1)
			enc.AddInt64("int64", 1)
			enc.AddFloat64("float64", 1.0)
			enc.AddString("string1", "\n")
			enc.AddString("string2", "💩")
			enc.AddString("string3", "🤔")
			enc.AddString("string4", "🙊")
			enc.AddBool("bool", true)
			enc.WriteEntry(ioutil.Discard, "fake", DebugLevel, ts)
			enc.Free()
		}
	})
}

func BenchmarkStandardJSON(b *testing.B) {
	record := logRecord{
		Level:   "debug",
		Message: "fake",
		Time:    time.Unix(0, 0),
		Fields: map[string]interface{}{
			"str":     "foo",
			"int":     int(1),
			"int64":   int64(1),
			"float64": float64(1.0),
			"string1": "\n",
			"string2": "💩",
			"string3": "🤔",
			"string4": "🙊",
			"bool":    true,
		},
	}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			json.Marshal(record)
		}
	})
}
