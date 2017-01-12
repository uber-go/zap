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
	"encoding/base64"
	"fmt"
	"math"
	"time"

	"go.uber.org/zap/internal/buffers"
	"go.uber.org/zap/zapcore"
)

// Skip constructs a no-op field.
func Skip() zapcore.Field {
	return zapcore.Field{Type: zapcore.SkipType}
}

// Base64 constructs a field that eagerly base64-encodes bytes.
func Base64(key string, val []byte) zapcore.Field {
	return String(key, base64.StdEncoding.EncodeToString(val))
}

// Bool constructs a field with the given key and value.
func Bool(key string, val bool) zapcore.Field {
	var ival int64
	if val {
		ival = 1
	}
	return zapcore.Field{Key: key, Type: zapcore.BoolType, Integer: ival}
}

// Float64 constructs a field with the given key and value. The way the
// floating-point value is represented is encoder-dependent, so marshaling is
// necessarily lazy.
func Float64(key string, val float64) zapcore.Field {
	return zapcore.Field{Key: key, Type: zapcore.Float64Type, Integer: int64(math.Float64bits(val))}
}

// Int constructs a field with the given key and value.
func Int(key string, val int) zapcore.Field {
	return Int64(key, int64(val))
}

// Int64 constructs a field with the given key and value.
func Int64(key string, val int64) zapcore.Field {
	return zapcore.Field{Key: key, Type: zapcore.Int64Type, Integer: val}
}

// Uint constructs a field with the given key and value.
func Uint(key string, val uint) zapcore.Field {
	return Uint64(key, uint64(val))
}

// Uint64 constructs a field with the given key and value.
func Uint64(key string, val uint64) zapcore.Field {
	return zapcore.Field{Key: key, Type: zapcore.Uint64Type, Integer: int64(val)}
}

// Uintptr constructs a field with the given key and value.
func Uintptr(key string, val uintptr) zapcore.Field {
	return Uint64(key, uint64(val))
}

// String constructs a field with the given key and value.
func String(key string, val string) zapcore.Field {
	return zapcore.Field{Key: key, Type: zapcore.StringType, String: val}
}

// Stringer constructs a field with the given key and the output of the value's
// String method. The Stringer's String method is called lazily.
func Stringer(key string, val fmt.Stringer) zapcore.Field {
	return zapcore.Field{Key: key, Type: zapcore.StringerType, Interface: val}
}

// Time constructs a zapcore.Field with the given key and value. It represents a
// time.Time as an integer number of milliseconds since epoch. Conversion to an
// int64 happens eagerly.
func Time(key string, val time.Time) zapcore.Field {
	return Int64(key, timeToMillis(val))
}

// Error constructs a field that lazily stores err.Error() under the key
// "error". If passed a nil error, the field is a no-op. This is purely a
// convenience for a common error-logging idiom; use String("someFieldName",
// err.Error()) to customize the key.
func Error(err error) zapcore.Field {
	if err == nil {
		return Skip()
	}
	return zapcore.Field{Key: "error", Type: zapcore.ErrorType, Interface: err}
}

// Stack constructs a field that stores a stacktrace of the current goroutine
// under the key "stacktrace". Keep in mind that taking a stacktrace is eager
// and extremely expensive (relatively speaking); this function both makes an
// allocation and takes ~10 microseconds.
func Stack() zapcore.Field {
	// Try to avoid allocating a buffer.
	buf := buffers.Get()
	// Returning the stacktrace as a string costs an allocation, but saves us
	// from expanding the zapcore.Field union struct to include a byte slice. Since
	// taking a stacktrace is already so expensive (~10us), the extra allocation
	// is okay.
	field := String("stacktrace", takeStacktrace(buf[:cap(buf)], false))
	buffers.Put(buf)
	return field
}

// Duration constructs a field with the given key and value. It represents
// durations as an integer number of nanoseconds.
func Duration(key string, val time.Duration) zapcore.Field {
	return Int64(key, int64(val))
}

// Object constructs a field with the given key and ObjectMarshaler. It
// provides a flexible, but still type-safe and efficient, way to add map- or
// struct-like user-defined types to the logging context. The struct's
// MarshalLogObject method is called lazily.
func Object(key string, val zapcore.ObjectMarshaler) zapcore.Field {
	return zapcore.Field{Key: key, Type: zapcore.ObjectMarshalerType, Interface: val}
}

// Array constructs a field with the given key and ArrayMarshaler. It provides
// a flexible, but still type-safe and efficient, way to add array-like types
// to the logging context. The struct's MarshalLogArray method is called lazily.
func Array(key string, val zapcore.ArrayMarshaler) zapcore.Field {
	return zapcore.Field{Key: key, Type: zapcore.ArrayMarshalerType, Interface: val}
}

// Reflect constructs a field with the given key and an arbitrary object. It uses
// an encoding-appropriate, reflection-based function to lazily serialize nearly
// any object into the logging context, but it's relatively slow and
// allocation-heavy.
//
// If encoding fails (e.g., trying to serialize a map[int]string to JSON), Reflect
// includes the error message in the final log output.
func Reflect(key string, val interface{}) zapcore.Field {
	return zapcore.Field{Key: key, Type: zapcore.ReflectType, Interface: val}
}

// Nest takes a key and a variadic number of zapcore.Fields and creates a nested
// namespace.
func Nest(key string, fields ...zapcore.Field) zapcore.Field {
	return zapcore.Field{Key: key, Type: zapcore.ObjectMarshalerType, Interface: zapcore.Fields(fields)}
}
