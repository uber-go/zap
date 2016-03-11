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

package benchmarks

import (
	"encoding/json"
	"io/ioutil"
	"testing"
	"time"

	"github.com/uber-common/zap/encoder"
)

type LogRecord struct {
	Level   string                 `json:"level"`
	Message string                 `json:"msg"`
	Time    time.Time              `json:"ts"`
	Fields  map[string]interface{} `json:"fields"`
}

func BenchmarkZapJSON(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			enc := encoder.NewJSON()
			enc.AddString("one", "foo")
			enc.AddInt("two", 1)
			enc.AddInt64("three", 1)
			enc.AddTime("four", time.Unix(0, 0))
			enc.AddFloat64("five", 1.0)
			enc.AddString("six", "\n")
			enc.AddString("seven", "💩")
			enc.AddString("eight", "🤔")
			enc.AddString("nine", "🙊")
			enc.AddBool("ten", true)
			enc.WriteMessage(ioutil.Discard, "debug", "fake", time.Unix(0, 0))
			enc.Free()
		}
	})
}

func BenchmarkStandardJSON(b *testing.B) {
	record := LogRecord{
		Level:   "debug",
		Message: "fake",
		Time:    time.Unix(0, 0),
		Fields: map[string]interface{}{
			"one":   "foo",
			"two":   int(1),
			"three": int64(1),
			"four":  time.Unix(0, 0),
			"five":  float64(1.0),
			"six":   "\n",
			"seven": "💩",
			"eight": "🤔",
			"nine":  "🙊",
			"ten":   true,
		},
	}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			json.Marshal(record)
		}
	})
}
