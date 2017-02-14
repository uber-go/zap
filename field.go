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
	"fmt"
	"math"
	"time"

	"go.uber.org/zap/internal/bufferpool"
	"go.uber.org/zap/zapcore"
)

// Skip constructs a no-op field.
func Skip() zapcore.Field {
	return zapcore.Field{Type: zapcore.SkipType}
}

// Binary constructs a field that carries an opaque binary blob.
//
// Binary data is serialized in an encoding-appropriate format. For example,
// zap's JSON encoder base64-encodes binary blobs.
func Binary(key string, val []byte) zapcore.Field {
	return zapcore.Field{Key: key, Type: zapcore.BinaryType, Interface: val}
}

// Bool constructs a field that carries a bool.
func Bool(key string, val bool) zapcore.Field {
	var ival int64
	if val {
		ival = 1
	}
	return zapcore.Field{Key: key, Type: zapcore.BoolType, Integer: ival}
}

// Complex128 constructs a field that carries a complex number. Unlike most
// numeric fields, this costs an allocation (to convert the complex128 to
// interface{}).
func Complex128(key string, val complex128) zapcore.Field {
	return zapcore.Field{Key: key, Type: zapcore.Complex128Type, Interface: val}
}

// Complex64 constructs a field that carries a complex number. Unlike most
// numeric fields, this costs an allocation (to convert the complex64 to
// interface{}).
func Complex64(key string, val complex64) zapcore.Field {
	return zapcore.Field{Key: key, Type: zapcore.Complex64Type, Interface: val}
}

// Float64 constructs a field that carries a float64. The way the
// floating-point value is represented is encoder-dependent, so marshaling is
// necessarily lazy.
func Float64(key string, val float64) zapcore.Field {
	return zapcore.Field{Key: key, Type: zapcore.Float64Type, Integer: int64(math.Float64bits(val))}
}

// Float32 constructs a field that carries a float32. The way the
// floating-point value is represented is encoder-dependent, so marshaling is
// necessarily lazy.
func Float32(key string, val float32) zapcore.Field {
	return zapcore.Field{Key: key, Type: zapcore.Float32Type, Integer: int64(math.Float32bits(val))}
}

// Int constructs a field with the given key and value.
func Int(key string, val int) zapcore.Field {
	return Int64(key, int64(val))
}

// Int64 constructs a field with the given key and value.
func Int64(key string, val int64) zapcore.Field {
	return zapcore.Field{Key: key, Type: zapcore.Int64Type, Integer: val}
}

// Int32 constructs a field with the given key and value.
func Int32(key string, val int32) zapcore.Field {
	return zapcore.Field{Key: key, Type: zapcore.Int32Type, Integer: int64(val)}
}

// Int16 constructs a field with the given key and value.
func Int16(key string, val int16) zapcore.Field {
	return zapcore.Field{Key: key, Type: zapcore.Int16Type, Integer: int64(val)}
}

// Int8 constructs a field with the given key and value.
func Int8(key string, val int8) zapcore.Field {
	return zapcore.Field{Key: key, Type: zapcore.Int8Type, Integer: int64(val)}
}

// String constructs a field with the given key and value.
func String(key string, val string) zapcore.Field {
	return zapcore.Field{Key: key, Type: zapcore.StringType, String: val}
}

// Uint constructs a field with the given key and value.
func Uint(key string, val uint) zapcore.Field {
	return Uint64(key, uint64(val))
}

// Uint64 constructs a field with the given key and value.
func Uint64(key string, val uint64) zapcore.Field {
	return zapcore.Field{Key: key, Type: zapcore.Uint64Type, Integer: int64(val)}
}

// Uint32 constructs a field with the given key and value.
func Uint32(key string, val uint32) zapcore.Field {
	return zapcore.Field{Key: key, Type: zapcore.Uint32Type, Integer: int64(val)}
}

// Uint16 constructs a field with the given key and value.
func Uint16(key string, val uint16) zapcore.Field {
	return zapcore.Field{Key: key, Type: zapcore.Uint16Type, Integer: int64(val)}
}

// Uint8 constructs a field with the given key and value.
func Uint8(key string, val uint8) zapcore.Field {
	return zapcore.Field{Key: key, Type: zapcore.Uint8Type, Integer: int64(val)}
}

// Uintptr constructs a field with the given key and value.
func Uintptr(key string, val uintptr) zapcore.Field {
	return zapcore.Field{Key: key, Type: zapcore.UintptrType, Integer: int64(val)}
}

// Reflect constructs a field with the given key and an arbitrary object. It uses
// an encoding-appropriate, reflection-based function to lazily serialize nearly
// any object into the logging context, but it's relatively slow and
// allocation-heavy. Outside tests, Any is always a better choice.
//
// If encoding fails (e.g., trying to serialize a map[int]string to JSON), Reflect
// includes the error message in the final log output.
func Reflect(key string, val interface{}) zapcore.Field {
	return zapcore.Field{Key: key, Type: zapcore.ReflectType, Interface: val}
}

// Namespace creates a named, isolated scope within the logger's context. All
// subsequent fields will be added to the new namespace.
//
// This helps prevent key collisions when injecting loggers into sub-components
// or third-party libraries.
func Namespace(key string) zapcore.Field {
	return zapcore.Field{Key: key, Type: zapcore.NamespaceType}
}

// Stringer constructs a field with the given key and the output of the value's
// String method. The Stringer's String method is called lazily.
func Stringer(key string, val fmt.Stringer) zapcore.Field {
	return zapcore.Field{Key: key, Type: zapcore.StringerType, Interface: val}
}

// Time constructs a zapcore.Field with the given key and value. The encoder
// controls how the time is serialized.
func Time(key string, val time.Time) zapcore.Field {
	return zapcore.Field{Key: key, Type: zapcore.TimeType, Integer: val.UnixNano()}
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
// under provided key. Keep in mind that taking a stacktrace is eager and
// extremely expensive (relatively speaking); this function both makes an
// allocation and takes ~10 microseconds.
func Stack(key string) zapcore.Field {
	// Try to avoid allocating a buffer.
	buf := bufferpool.Get()
	bs := buf.Bytes()
	// Returning the stacktrace as a string costs an allocation, but saves us
	// from expanding the zapcore.Field union struct to include a byte slice. Since
	// taking a stacktrace is already so expensive (~10us), the extra allocation
	// is okay.
	field := String(key, takeStacktrace(bs[:cap(bs)], false))
	bufferpool.Put(buf)
	return field
}

// Duration constructs a field with the given key and value. The encoder
// controls how the duration is serialized.
func Duration(key string, val time.Duration) zapcore.Field {
	return zapcore.Field{Key: key, Type: zapcore.DurationType, Integer: int64(val)}
}

// Object constructs a field with the given key and ObjectMarshaler. It
// provides a flexible, but still type-safe and efficient, way to add map- or
// struct-like user-defined types to the logging context. The struct's
// MarshalLogObject method is called lazily.
func Object(key string, val zapcore.ObjectMarshaler) zapcore.Field {
	return zapcore.Field{Key: key, Type: zapcore.ObjectMarshalerType, Interface: val}
}

// Any takes a key and an arbitrary value and chooses the best way to represent
// them as a field, falling back to a reflection-based approach only if
// necessary.
func Any(key string, value interface{}) zapcore.Field {
	switch val := value.(type) {
	case zapcore.ObjectMarshaler:
		return Object(key, val)
	case zapcore.ArrayMarshaler:
		return Array(key, val)
	case bool:
		return Bool(key, val)
	case []bool:
		return Bools(key, val)
	case complex128:
		return Complex128(key, val)
	case []complex128:
		return Complex128s(key, val)
	case complex64:
		return Complex64(key, val)
	case []complex64:
		return Complex64s(key, val)
	case float64:
		return Float64(key, val)
	case []float64:
		return Float64s(key, val)
	case float32:
		return Float32(key, val)
	case []float32:
		return Float32s(key, val)
	case int:
		return Int(key, val)
	case []int:
		return Ints(key, val)
	case int64:
		return Int64(key, val)
	case []int64:
		return Int64s(key, val)
	case int32:
		return Int32(key, val)
	case []int32:
		return Int32s(key, val)
	case int16:
		return Int16(key, val)
	case []int16:
		return Int16s(key, val)
	case int8:
		return Int8(key, val)
	case []int8:
		return Int8s(key, val)
	case string:
		return String(key, val)
	case []string:
		return Strings(key, val)
	case uint:
		return Uint(key, val)
	case []uint:
		return Uints(key, val)
	case uint64:
		return Uint64(key, val)
	case []uint64:
		return Uint64s(key, val)
	case uint32:
		return Uint32(key, val)
	case []uint32:
		return Uint32s(key, val)
	case uint16:
		return Uint16(key, val)
	case []uint16:
		return Uint16s(key, val)
	case uint8:
		return Uint8(key, val)
	case []uint8:
		return Uint8s(key, val)
	case uintptr:
		return Uintptr(key, val)
	case []uintptr:
		return Uintptrs(key, val)
	case time.Time:
		return Time(key, val)
	case []time.Time:
		return Times(key, val)
	case time.Duration:
		return Duration(key, val)
	case []time.Duration:
		return Durations(key, val)
	case error:
		return String(key, val.Error())
	case []error:
		return Errors(key, val)
	case fmt.Stringer:
		return Stringer(key, val)
	default:
		return Reflect(key, val)
	}
}
