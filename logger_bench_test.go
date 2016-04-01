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

package zap_test

import (
	"errors"
	"io/ioutil"
	"strconv"
	"testing"
	"time"

	"github.com/uber-common/zap"
)

type user struct {
	name      string
	email     string
	createdAt time.Time
}

func (u user) MarshalLog(kv zap.KeyValue) error {
	kv.AddString("name", u.name)
	kv.AddString("email", u.email)
	kv.AddTime("created_at", u.createdAt)
	return nil
}

var _jane = user{
	name:      "Jane Doe",
	email:     "jane@test.com",
	createdAt: time.Date(1980, 1, 1, 12, 0, 0, 0, time.UTC),
}

func benchField(b *testing.B, fields ...zap.Field) {
	logger := zap.NewJSON(zap.All, ioutil.Discard)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info("Go fast.", fields...)
		}
	})
}

func benchManyFields(b *testing.B, numFields int) {
	fields := make([]zap.Field, numFields)
	for i := 0; i < numFields; i++ {
		fields[i] = zap.Int(strconv.Itoa(i), i)
	}
	benchField(b, fields...)
}

func BenchmarkNoContext(b *testing.B)     { benchField(b) }
func BenchmarkBoolField(b *testing.B)     { benchField(b, zap.Bool("foo", true)) }
func BenchmarkFloat64Field(b *testing.B)  { benchField(b, zap.Float64("foo", 3.14)) }
func BenchmarkIntField(b *testing.B)      { benchField(b, zap.Int("foo", 42)) }
func BenchmarkInt64Field(b *testing.B)    { benchField(b, zap.Int64("foo", 42)) }
func BenchmarkStringField(b *testing.B)   { benchField(b, zap.String("foo", "bar")) }
func BenchmarkTimeField(b *testing.B)     { benchField(b, zap.Time("foo", time.Unix(0, 0))) }
func BenchmarkDurationField(b *testing.B) { benchField(b, zap.Duration("foo", time.Second)) }
func BenchmarkErrField(b *testing.B)      { benchField(b, zap.Err(errors.New("egad!"))) }
func BenchmarkObjectField(b *testing.B)   { benchField(b, zap.Object("user", _jane)) }
func Benchmark10Fields(b *testing.B)      { benchManyFields(b, 10) }
func Benchmark50Fields(b *testing.B)      { benchManyFields(b, 50) }
func Benchmark100Fields(b *testing.B)     { benchManyFields(b, 100) }
func Benchmark200Fields(b *testing.B)     { benchManyFields(b, 200) }
func Benchmark500Fields(b *testing.B)     { benchManyFields(b, 500) }
