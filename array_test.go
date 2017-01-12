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
	"testing"

	"github.com/stretchr/testify/assert"

	"go.uber.org/zap/zapcore"
)

func BenchmarkBoolsArrayMarshaler(b *testing.B) {
	// Keep this benchmark here to capture the overhead of the ArrayMarshaler
	// wrapper.
	bs := make([]bool, 50)
	enc := zapcore.NewJSONEncoder(zapcore.JSONConfig{})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Array("array", Bools(bs)).AddTo(enc.Clone())
	}
}

func BenchmarkBoolsReflect(b *testing.B) {
	bs := make([]bool, 50)
	enc := zapcore.NewJSONEncoder(zapcore.JSONConfig{})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Reflect("array", bs).AddTo(enc.Clone())
	}
}

func TestArrayWrappers(t *testing.T) {
	tests := []struct {
		desc     string
		array    zapcore.ArrayMarshaler
		expected []interface{}
	}{
		{
			"arrays",
			Arrays([]zapcore.ArrayMarshaler{
				Bools([]bool{true}),
				Bools([]bool{false}),
			}),
			[]interface{}{
				[]interface{}{true},
				[]interface{}{false},
			},
		},
		{
			"objects",
			Objects([]zapcore.ObjectMarshaler{
				zapcore.ObjectMarshalerFunc(func(_ zapcore.ObjectEncoder) error { return nil }),
				zapcore.ObjectMarshalerFunc(func(_ zapcore.ObjectEncoder) error { return nil }),
			}),
			[]interface{}{zapcore.MapObjectEncoder{}, zapcore.MapObjectEncoder{}},
		},
		{"bools", Bools([]bool{true, false}), []interface{}{true, false}},
	}

	for _, tt := range tests {
		enc := make(zapcore.MapObjectEncoder)
		Array("k", tt.array).AddTo(enc)
		assert.Equal(t, tt.expected, enc["k"], "%s: unexpected map contents.", tt.desc)
		assert.Equal(t, 1, len(enc), "%s: found extra keys in map: %v", tt.desc, enc)
	}
}
