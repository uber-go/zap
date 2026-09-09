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

package buffer

import (
	"bytes"
	"math"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBufferWrites(t *testing.T) {
	buf := NewPool().Get()

	tests := []struct {
		desc string
		f    func()
		want string
	}{
		{"AppendByte", func() { buf.AppendByte('v') }, "v"},
		{"AppendString", func() { buf.AppendString("foo") }, "foo"},
		{"AppendIntPositive", func() { buf.AppendInt(42) }, "42"},
		{"AppendIntNegative", func() { buf.AppendInt(-42) }, "-42"},
		{"AppendUint", func() { buf.AppendUint(42) }, "42"},
		{"AppendBool", func() { buf.AppendBool(true) }, "true"},
		{"AppendFloat64", func() { buf.AppendFloat(3.14, 64) }, "3.14"},
		// Intentionally introduce some floating-point error.
		{"AppendFloat32", func() { buf.AppendFloat(float64(float32(3.14)), 32) }, "3.14"},
		{"AppendNumberZeroScale", func() { buf.AppendNumber(42, 0) }, "42"},
		{"AppendNumberNegativeZeroScale", func() { buf.AppendNumber(-42, 0) }, "-42"},
		{"AppendNumberNegativeScale", func() { buf.AppendNumber(42, -2) }, "4200"},
		{"AppendNumberNegativeValueNegativeScale", func() { buf.AppendNumber(-42, -2) }, "-4200"},
		{"AppendNumberZeroNegativeScale", func() { buf.AppendNumber(0, -2) }, "0"},
		{"AppendNumberMinInt64NegativeScale", func() { buf.AppendNumber(-9223372036854775808, -2) }, "-922337203685477580800"},
		{"AppendNumberMaxInt64NegativeScale", func() { buf.AppendNumber(9223372036854775807, -2) }, "922337203685477580700"},
		{"AppendNumberRoundSeconds", func() { buf.AppendNumber(1008720000000000000, 9) }, "1008720000"},
		{"AppendNumberRoundMillis", func() { buf.AppendNumber(1008720000000000000, 6) }, "1008720000000"},
		{"AppendNumberFullNanos", func() { buf.AppendNumber(1735689600123456789, 9) }, "1735689600.123456789"},
		{"AppendNumberFullMillisRemainder", func() { buf.AppendNumber(1735689600123456789, 6) }, "1735689600123.456789"},
		{"AppendNumberStripsTrailingZeros", func() { buf.AppendNumber(1735689600123456000, 6) }, "1735689600123.456"},
		{"AppendNumberSubOne", func() { buf.AppendNumber(123, 6) }, "0.000123"},
		{"AppendNumberSubOneNoLeadingZeros", func() { buf.AppendNumber(123, 3) }, "0.123"},
		{"AppendNumberSubOneOverlap", func() { buf.AppendNumber(123456, 6) }, "0.123456"},
		{"AppendNumberSubOneStripped", func() { buf.AppendNumber(100, 6) }, "0.0001"},
		{"AppendNumberZero", func() { buf.AppendNumber(0, 9) }, "0"},
		{"AppendNumberMinInt64", func() { buf.AppendNumber(-9223372036854775808, 9) }, "-9223372036.854775808"},
		{"AppendWrite", func() { buf.Write([]byte("foo")) }, "foo"},
		{"AppendTime", func() { buf.AppendTime(time.Date(2000, 1, 2, 3, 4, 5, 6, time.UTC), time.RFC3339) }, "2000-01-02T03:04:05Z"},
		{"WriteByte", func() { buf.WriteByte('v') }, "v"},
		{"WriteString", func() { buf.WriteString("foo") }, "foo"},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			buf.Reset()
			tt.f()
			assert.Equal(t, tt.want, buf.String(), "Unexpected buffer.String().")
			assert.Equal(t, tt.want, string(buf.Bytes()), "Unexpected string(buffer.Bytes()).")
			assert.Equal(t, len(tt.want), buf.Len(), "Unexpected buffer length.")
			// We're not writing more than a kibibyte in tests.
			assert.Equal(t, _size, buf.Cap(), "Expected buffer capacity to remain constant.")
		})
	}
}

func TestAppendNumberGrowth(t *testing.T) {
	const prefix = "prefix:"
	for _, tt := range []struct {
		name  string
		n     int64
		scale int
		want  string
	}{
		{"decimal", 42, 1, "4.2"},
		{"negativeDecimal", -42, 1, "-4.2"},
		{"subOne", 42, 4, "0.0042"},
		{"negativeSubOne", -42, 4, "-0.0042"},
		{"negativeScale", 42, -2, "4200"},
		{"negativeValueNegativeScale", -42, -2, "-4200"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			// Fit the integer exactly so decimal formatting must grow the buffer.
			capacity := len(prefix) + len(strconv.FormatInt(tt.n, 10))
			buf := &Buffer{bs: make([]byte, 0, capacity)}
			buf.AppendString(prefix)
			buf.AppendNumber(tt.n, tt.scale)
			assert.Equal(t, prefix+tt.want, buf.String())
			assert.Greater(t, buf.Cap(), capacity)
		})
	}
}

func TestAppendNumberAllocs(t *testing.T) {
	for _, tt := range []struct {
		name  string
		n     int64
		scale int
	}{
		{"zero", 0, 9},
		{"integer", 42, 0},
		{"negativeScale", -42, -2},
		{"subOne", -123, 6},
		{"seconds", 1735689600123456789, 9},
		{"millis", 1735689600123456789, 6},
		{"trailingZeros", 1735689600000000000, 9},
	} {
		t.Run(tt.name, func(t *testing.T) {
			buf := NewPool().Get()
			defer buf.Free()
			assert.Zero(t, testing.AllocsPerRun(100, func() {
				buf.Reset()
				buf.AppendNumber(tt.n, tt.scale)
			}))
		})
	}
}

func BenchmarkBuffers(b *testing.B) {
	// Because we use the strconv.AppendFoo functions so liberally, we can't
	// use the standard library's bytes.Buffer anyways (without incurring a
	// bunch of extra allocations). Nevertheless, let's make sure that we're
	// not losing any precious nanoseconds.
	str := strings.Repeat("a", 1024)
	slice := make([]byte, 0, 1024)
	buf := bytes.NewBuffer(slice)
	custom := NewPool().Get()
	b.Run("ByteSlice", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			slice = append(slice, str...)
			slice = slice[:0]
		}
	})
	b.Run("BytesBuffer", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			buf.WriteString(str)
			buf.Reset()
		}
	})
	b.Run("CustomBuffer", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			custom.AppendString(str)
			custom.Reset()
		}
	})
}

func BenchmarkAppendNumber(b *testing.B) {
	tests := []struct {
		name  string
		n     int64
		scale int
	}{
		{"subOne", 123, 6},
		{"negativeSubOne", -123, 6},
		{"fullMillis", 1735689600123456789, 6},
		{"fullSeconds", 1735689600123456789, 9},
		{"trailingZeros", 1735689600000000000, 6},
		{"negativeScale", -42, -2},
		{"largeScale", 1, 24},
		{"zero", 0, 9},
		{"integer", 42, 0},
	}
	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			buf := NewPool().Get()
			defer buf.Free()
			b.Run("scale", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					buf.Reset()
					buf.AppendNumber(tt.n, tt.scale)
				}
			})
			// Isolate formatting, using the same scaled input as AppendNumber.
			value := float64(tt.n) / math.Pow10(tt.scale)
			b.Run("float", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					buf.Reset()
					buf.AppendFloat(value, 64)
				}
			})
		})
	}
}
