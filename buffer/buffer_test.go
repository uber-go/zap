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
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBufferWrites(t *testing.T) {
	t.Parallel()

	tests := []struct {
		desc string
		f    func(*Buffer)
		want string
	}{
		{
			desc: "AppendByte",
			f:    func(buf *Buffer) { buf.AppendByte('v') },
			want: "v",
		},
		{
			desc: "AppendString",
			f:    func(buf *Buffer) { buf.AppendString("foo") },
			want: "foo",
		},
		{
			desc: "AppendIntPositive",
			f:    func(buf *Buffer) { buf.AppendInt(42) },
			want: "42",
		},
		{
			desc: "AppendIntNegative",
			f:    func(buf *Buffer) { buf.AppendInt(-42) },
			want: "-42",
		},
		{
			desc: "AppendUint",
			f:    func(buf *Buffer) { buf.AppendUint(42) },
			want: "42",
		},
		{
			desc: "AppendBool",
			f:    func(buf *Buffer) { buf.AppendBool(true) },
			want: "true",
		},
		{
			desc: "AppendFloat64",
			f:    func(buf *Buffer) { buf.AppendFloat(3.14, 64) },
			want: "3.14",
		},
		// Intentionally introduce some floating-point error.
		{
			desc: "AppendFloat32",
			f:    func(buf *Buffer) { buf.AppendFloat(float64(float32(3.14)), 32) },
			want: "3.14",
		},
		{
			desc: "AppendWrite",
			f:    func(buf *Buffer) { buf.Write([]byte("foo")) },
			want: "foo",
		},
		{
			desc: "AppendTime",
			f:    func(buf *Buffer) { buf.AppendTime(time.Date(2000, 1, 2, 3, 4, 5, 6, time.UTC), time.RFC3339) },
			want: "2000-01-02T03:04:05Z",
		},
		{
			desc: "WriteByte",
			f:    func(buf *Buffer) { buf.WriteByte('v') },
			want: "v",
		},
		{
			desc: "WriteString",
			f:    func(buf *Buffer) { buf.WriteString("foo") },
			want: "foo",
		},
	}

	pool := NewPool()
	for _, tt := range tests {
		tt := tt
		t.Run(tt.desc, func(t *testing.T) {
			t.Parallel()

			buf := pool.Get()
			defer buf.Free()

			tt.f(buf)
			assert.Equal(t, tt.want, buf.String(), "Unexpected buffer.String().")
			assert.Equal(t, tt.want, string(buf.Bytes()), "Unexpected string(buffer.Bytes()).")
			assert.Equal(t, len(tt.want), buf.Len(), "Unexpected buffer length.")
			// We're not writing more than a kibibyte in tests.
			assert.Equal(t, _size, buf.Cap(), "Expected buffer capacity to remain constant.")
		})
	}
}

func BenchmarkBuffers(b *testing.B) {
	// Because we use the strconv.AppendFoo functions so liberally, we can't
	// use the standard library's bytes.Buffer anyways (without incurring a
	// bunch of extra allocations). Nevertheless, let's make sure that we're
	// not losing any precious nanoseconds.
	str := strings.Repeat("a", 1024)
	slice := make([]byte, 1024)
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
