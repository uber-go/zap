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
	"io/ioutil"
	"testing"
	"time"
)

func BenchmarkNullLogMarshalerFunc(b *testing.B) {
	for i := 0; i < b.N; i++ {
		enc := NullEncoder()
		enc.AddMarshaler("nested", LogMarshalerFunc(func(kv KeyValue) error {
			kv.AddInt("i", i)
			return nil
		}))
		enc.Free()
	}
}

func BenchmarkZapNull(b *testing.B) {
	ts := time.Unix(0, 0)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			enc := NullEncoder()
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
