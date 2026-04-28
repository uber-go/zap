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
	"errors"
	"fmt"
	"testing"
	"time"

	"go.uber.org/zap/zapcore"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func BenchmarkBoolsArrayMarshaler(b *testing.B) {
	// Keep this benchmark here to capture the overhead of the ArrayMarshaler
	// wrapper.
	bs := make([]bool, 50)
	enc := zapcore.NewJSONEncoder(zapcore.EncoderConfig{})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Bools("array", bs).AddTo(enc.Clone())
	}
}

func BenchmarkBoolsReflect(b *testing.B) {
	bs := make([]bool, 50)
	enc := zapcore.NewJSONEncoder(zapcore.EncoderConfig{})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Reflect("array", bs).AddTo(enc.Clone())
	}
}

func TestArrayWrappers(t *testing.T) {
	tests := []struct {
		desc     string
		field    Field
		expected []interface{}
	}{
		{"empty bools", Bools("", []bool{}), []interface{}{}},
		{"empty byte strings", ByteStrings("", [][]byte{}), []interface{}{}},
		{"empty complex128s", Complex128s("", []complex128{}), []interface{}{}},
		{"empty complex64s", Complex64s("", []complex64{}), []interface{}{}},
		{"empty durations", Durations("", []time.Duration{}), []interface{}{}},
		{"empty float64s", Float64s("", []float64{}), []interface{}{}},
		{"empty float32s", Float32s("", []float32{}), []interface{}{}},
		{"empty ints", Ints("", []int{}), []interface{}{}},
		{"empty int64s", Int64s("", []int64{}), []interface{}{}},
		{"empty int32s", Int32s("", []int32{}), []interface{}{}},
		{"empty int16s", Int16s("", []int16{}), []interface{}{}},
		{"empty int8s", Int8s("", []int8{}), []interface{}{}},
		{"empty strings", Strings("", []string{}), []interface{}{}},
		{"empty times", Times("", []time.Time{}), []interface{}{}},
		{"empty uints", Uints("", []uint{}), []interface{}{}},
		{"empty uint64s", Uint64s("", []uint64{}), []interface{}{}},
		{"empty uint32s", Uint32s("", []uint32{}), []interface{}{}},
		{"empty uint16s", Uint16s("", []uint16{}), []interface{}{}},
		{"empty uint8s", Uint8s("", []uint8{}), []interface{}{}},
		{"empty uintptrs", Uintptrs("", []uintptr{}), []interface{}{}},
		{"bools", Bools("", []bool{true, false}), []interface{}{true, false}},
		{"byte strings", ByteStrings("", [][]byte{{1, 2}, {3, 4}}), []interface{}{"\x01\x02", "\x03\x04"}},
		{"complex128s", Complex128s("", []complex128{1 + 2i, 3 + 4i}), []interface{}{1 + 2i, 3 + 4i}},
		{"complex64s", Complex64s("", []complex64{1 + 2i, 3 + 4i}), []interface{}{complex64(1 + 2i), complex64(3 + 4i)}},
		{"durations", Durations("", []time.Duration{1, 2}), []interface{}{time.Nanosecond, 2 * time.Nanosecond}},
		{"float64s", Float64s("", []float64{1.2, 3.4}), []interface{}{1.2, 3.4}},
		{"float32s", Float32s("", []float32{1.2, 3.4}), []interface{}{float32(1.2), float32(3.4)}},
		{"ints", Ints("", []int{1, 2}), []interface{}{1, 2}},
		{"int64s", Int64s("", []int64{1, 2}), []interface{}{int64(1), int64(2)}},
		{"int32s", Int32s("", []int32{1, 2}), []interface{}{int32(1), int32(2)}},
		{"int16s", Int16s("", []int16{1, 2}), []interface{}{int16(1), int16(2)}},
		{"int8s", Int8s("", []int8{1, 2}), []interface{}{int8(1), int8(2)}},
		{"strings", Strings("", []string{"foo", "bar"}), []interface{}{"foo", "bar"}},
		{"times", Times("", []time.Time{time.Unix(0, 0), time.Unix(0, 0)}), []interface{}{time.Unix(0, 0), time.Unix(0, 0)}},
		{"uints", Uints("", []uint{1, 2}), []interface{}{uint(1), uint(2)}},
		{"uint64s", Uint64s("", []uint64{1, 2}), []interface{}{uint64(1), uint64(2)}},
		{"uint32s", Uint32s("", []uint32{1, 2}), []interface{}{uint32(1), uint32(2)}},
		{"uint16s", Uint16s("", []uint16{1, 2}), []interface{}{uint16(1), uint16(2)}},
		{"uint8s", Uint8s("", []uint8{1, 2}), []interface{}{uint8(1), uint8(2)}},
		{"uintptrs", Uintptrs("", []uintptr{1, 2}), []interface{}{uintptr(1), uintptr(2)}},
	}

	for _, tt := range tests {
		enc := zapcore.NewMapObjectEncoder()
		tt.field.Key = "k"
		tt.field.AddTo(enc)
		assert.Equal(t, tt.expected, enc.Fields["k"], "%s: unexpected map contents.", tt.desc)
		assert.Equal(t, 1, len(enc.Fields), "%s: found extra keys in map: %v", tt.desc, enc.Fields)
	}
}

func TestObjectsAndObjectValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		desc string
		give Field
		want []any
	}{
		{
			desc: "Objects/nil slice",
			give: Objects[*emptyObject]("", nil),
			want: []any{},
		},
		{
			desc: "ObjectValues/nil slice",
			give: ObjectValues[emptyObject]("", nil),
			want: []any{},
		},
		{
			desc: "ObjectValues/empty slice",
			give: ObjectValues("", []emptyObject{}),
			want: []any{},
		},
		{
			desc: "ObjectValues/single item",
			give: ObjectValues("", []emptyObject{
				{},
			}),
			want: []any{
				map[string]any{},
			},
		},
		{
			desc: "Objects/multiple different objects",
			give: Objects("", []*fakeObject{
				{value: "foo"},
				{value: "bar"},
				{value: "baz"},
			}),
			want: []any{
				map[string]any{"value": "foo"},
				map[string]any{"value": "bar"},
				map[string]any{"value": "baz"},
			},
		},
		{
			desc: "ObjectValues/multiple different objects",
			give: ObjectValues("", []fakeObject{
				{value: "foo"},
				{value: "bar"},
				{value: "baz"},
			}),
			want: []any{
				map[string]any{"value": "foo"},
				map[string]any{"value": "bar"},
				map[string]any{"value": "baz"},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.desc, func(t *testing.T) {
			t.Parallel()

			tt.give.Key = "k"

			enc := zapcore.NewMapObjectEncoder()
			tt.give.AddTo(enc)
			assert.Equal(t, tt.want, enc.Fields["k"])
		})
	}
}

type emptyObject struct{}

func (*emptyObject) MarshalLogObject(zapcore.ObjectEncoder) error {
	return nil
}

type fakeObject struct {
	value string
	err   error // marshaling error, if any
}

func (o *fakeObject) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("value", o.value)
	return o.err
}

func TestObjectsAndObjectValues_marshalError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		desc    string
		give    Field
		want    []any
		wantErr string
	}{
		{
			desc: "Objects",
			give: Objects("", []*fakeObject{
				{value: "foo"},
				{value: "bar", err: errors.New("great sadness")},
				{value: "baz"}, // does not get marshaled
			}),
			want: []any{
				map[string]any{"value": "foo"},
				map[string]any{"value": "bar"},
			},
			wantErr: "great sadness",
		},
		{
			desc: "ObjectValues",
			give: ObjectValues("", []fakeObject{
				{value: "foo"},
				{value: "bar", err: errors.New("stuff failed")},
				{value: "baz"}, // does not get marshaled
			}),
			want: []any{
				map[string]any{"value": "foo"},
				map[string]any{"value": "bar"},
			},
			wantErr: "stuff failed",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.desc, func(t *testing.T) {
			t.Parallel()

			tt.give.Key = "k"

			enc := zapcore.NewMapObjectEncoder()
			tt.give.AddTo(enc)

			require.Contains(t, enc.Fields, "k")
			assert.Equal(t, tt.want, enc.Fields["k"])

			// AddTo puts the error in a "%vError" field based on the name of the
			// original field.
			require.Contains(t, enc.Fields, "kError")
			assert.Equal(t, tt.wantErr, enc.Fields["kError"])
		})
	}
}

type stringerObject struct {
	value string
}

func (s stringerObject) String() string {
	return s.value
}

func TestStringers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		desc string
		give Field
		want []any
	}{
		{
			desc: "Stringers",
			give: Stringers("", []stringerObject{
				{value: "foo"},
				{value: "bar"},
				{value: "baz"},
			}),
			want: []any{
				"foo",
				"bar",
				"baz",
			},
		},
		{
			desc: "Stringers with []fmt.Stringer",
			give: Stringers("", []fmt.Stringer{
				stringerObject{value: "foo"},
				stringerObject{value: "bar"},
				stringerObject{value: "baz"},
			}),
			want: []any{
				"foo",
				"bar",
				"baz",
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.desc, func(t *testing.T) {
			t.Parallel()

			tt.give.Key = "k"

			enc := zapcore.NewMapObjectEncoder()
			tt.give.AddTo(enc)
			assert.Equal(t, tt.want, enc.Fields["k"])
		})
	}
}
