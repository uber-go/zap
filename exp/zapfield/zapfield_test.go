// Copyright (c) 2016-2023 Uber Technologies, Inc.
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

package zapfield

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type (
	EmbeddedStringKey string

	EmbeddedUint   uint
	EmbeddedUint64 uint64
	EmbeddedUint32 uint32
	EmbeddedUint16 uint16
	EmbeddedUint8  uint8

	EmbeddedInt   int
	EmbeddedInt64 int64
	EmbeddedInt32 int32
	EmbeddedInt16 int16
	EmbeddedInt8  int8

	EmbeddedFloat64 float64
	EmbeddedFloat32 float32

	EmbeddedString  string
	EmbeddedStrings []EmbeddedString
)

func TestFieldConstructors(t *testing.T) {
	var (
		key = EmbeddedStringKey("test key")

		uintValue   = EmbeddedUint(1)
		uint64Value = EmbeddedUint64(2)
		uint32Value = EmbeddedUint32(3)
		uint16Value = EmbeddedUint16(4)
		uint8Value  = EmbeddedUint8(5)

		intValue   = EmbeddedInt(-1)
		int64Value = EmbeddedInt64(20)
		int32Value = EmbeddedInt32(-300)
		int16Value = EmbeddedInt16(4000)
		int8Value  = EmbeddedInt8(-54)

		float64Value = EmbeddedFloat64(123.45)
		float32Value = EmbeddedFloat32(-54.321)

		stringValue  = EmbeddedString("test value")
		stringsValue = EmbeddedStrings{
			EmbeddedString("test value 1"),
			EmbeddedString("test value 2"),
		}
	)

	tests := []struct {
		name   string
		expect zap.Field
		field  zap.Field
	}{
		{"Uint", zap.Field{Type: zapcore.Uint64Type, Key: "test key", Integer: 1}, Uint(key, uintValue)},
		{"Uint64", zap.Field{Type: zapcore.Uint64Type, Key: "test key", Integer: 2}, Uint64(key, uint64Value)},
		{"Uint32", zap.Field{Type: zapcore.Uint32Type, Key: "test key", Integer: 3}, Uint32(key, uint32Value)},
		{"Uint16", zap.Field{Type: zapcore.Uint16Type, Key: "test key", Integer: 4}, Uint16(key, uint16Value)},
		{"Uint8", zap.Field{Type: zapcore.Uint8Type, Key: "test key", Integer: 5}, Uint8(key, uint8Value)},

		{"Int", zap.Field{Type: zapcore.Int64Type, Key: "test key", Integer: -1}, Int(key, intValue)},
		{"Int64", zap.Field{Type: zapcore.Int64Type, Key: "test key", Integer: 20}, Int64(key, int64Value)},
		{"Int32", zap.Field{Type: zapcore.Int32Type, Key: "test key", Integer: -300}, Int32(key, int32Value)},
		{"Int16", zap.Field{Type: zapcore.Int16Type, Key: "test key", Integer: 4000}, Int16(key, int16Value)},
		{"Int8", zap.Field{Type: zapcore.Int8Type, Key: "test key", Integer: -54}, Int8(key, int8Value)},

		{"Float64", zap.Field{Type: zapcore.Float64Type, Key: "test key", Integer: 4638387438405602509}, Float64(key, float64Value)},
		{"Float32", zap.Field{Type: zapcore.Float32Type, Key: "test key", Integer: 3260631220}, Float32(key, float32Value)},

		{"String", zap.Field{Type: zapcore.StringType, Key: "test key", String: "test value"}, String(key, stringValue)},
		{"Str", zap.Field{Type: zapcore.StringType, Key: "test key", String: "test value"}, Str(key, stringValue)},
		{"Strings", zap.Array("test key", stringArray[EmbeddedString]{"test value 1", "test value 2"}), Strings(key, stringsValue)},
		{"Strs", zap.Array("test key", stringArray[EmbeddedString]{"test value 1", "test value 2"}), Strs(key, stringsValue)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !assert.Equal(t, tt.expect, tt.field, "Unexpected output from convenience field constructor %s.", tt.name) {
				t.Logf("type expected: %T\nGot: %T", tt.expect.Interface, tt.field.Interface)
			}
			assertCanBeReused(t, tt.field)
		})
	}
}

func assertCanBeReused(t testing.TB, field zap.Field) {
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		enc := zapcore.NewMapObjectEncoder()

		// Ensure using the field in multiple encoders in separate goroutines
		// does not cause any races or panics.
		wg.Add(1)
		go func() {
			defer wg.Done()
			assert.NotPanics(t, func() {
				field.AddTo(enc)
			}, "Reusing a field should not cause issues")
		}()
	}

	wg.Wait()
}
