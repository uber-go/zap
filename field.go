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
	"unsafe"

	"go.uber.org/zap/zapcore"
)

// Field is an alias for Field. Aliasing this type dramatically
// improves the navigability of this package's API documentation.
type Field = zapcore.Field

var (
	_minTimeInt64 = time.Unix(0, math.MinInt64)
	_maxTimeInt64 = time.Unix(0, math.MaxInt64)
)

// Skip constructs a no-op field, which is often useful when handling invalid
// inputs in other Field constructors.
func Skip() Field {
	return Field{Type: zapcore.SkipType}
}

// nilField returns a field which will marshal explicitly as nil. See motivation
// in https://github.com/uber-go/zap/issues/753 . If we ever make breaking
// changes and add zapcore.NilType and zapcore.ObjectEncoder.AddNil, the
// implementation here should be changed to reflect that.
func nilField(key string) Field { return Reflect(key, nil) }

// Binary constructs a field that carries an opaque binary blob.
//
// Binary data is serialized in an encoding-appropriate format. For example,
// zap's JSON encoder base64-encodes binary blobs. To log UTF-8 encoded text,
// use ByteString.
func Binary(key string, val []byte) Field {
	return Field{Key: key, Type: zapcore.BinaryType, Interface: val}
}

// Bool constructs a field that carries a bool.
func Bool(key string, val bool) Field {
	var ival int64
	if val {
		ival = 1
	}
	return Field{Key: key, Type: zapcore.BoolType, Integer: ival}
}

// Boolp constructs a field that carries a *bool. The returned Field will safely
// and explicitly represent `nil` when appropriate.
func Boolp(key string, val *bool) Field {
	if val == nil {
		return nilField(key)
	}
	return Bool(key, *val)
}

// ByteString constructs a field that carries UTF-8 encoded text as a []byte.
// To log opaque binary blobs (which aren't necessarily valid UTF-8), use
// Binary.
func ByteString(key string, val []byte) Field {
	return Field{Key: key, Type: zapcore.ByteStringType, Interface: val}
}

// Complex128 constructs a field that carries a complex number. Unlike most
// numeric fields, this costs an allocation (to convert the complex128 to
// interface{}).
func Complex128(key string, val complex128) Field {
	return Field{Key: key, Type: zapcore.Complex128Type, Interface: val}
}

// Complex128p constructs a field that carries a *complex128. The returned Field will safely
// and explicitly represent `nil` when appropriate.
func Complex128p(key string, val *complex128) Field {
	if val == nil {
		return nilField(key)
	}
	return Complex128(key, *val)
}

// Complex64 constructs a field that carries a complex number. Unlike most
// numeric fields, this costs an allocation (to convert the complex64 to
// interface{}).
func Complex64(key string, val complex64) Field {
	return Field{Key: key, Type: zapcore.Complex64Type, Interface: val}
}

// Complex64p constructs a field that carries a *complex64. The returned Field will safely
// and explicitly represent `nil` when appropriate.
func Complex64p(key string, val *complex64) Field {
	if val == nil {
		return nilField(key)
	}
	return Complex64(key, *val)
}

// Float64 constructs a field that carries a float64. The way the
// floating-point value is represented is encoder-dependent, so marshaling is
// necessarily lazy.
func Float64(key string, val float64) Field {
	return Field{Key: key, Type: zapcore.Float64Type, Integer: int64(math.Float64bits(val))}
}

// Float64p constructs a field that carries a *float64. The returned Field will safely
// and explicitly represent `nil` when appropriate.
func Float64p(key string, val *float64) Field {
	if val == nil {
		return nilField(key)
	}
	return Float64(key, *val)
}

// Float32 constructs a field that carries a float32. The way the
// floating-point value is represented is encoder-dependent, so marshaling is
// necessarily lazy.
func Float32(key string, val float32) Field {
	return Field{Key: key, Type: zapcore.Float32Type, Integer: int64(math.Float32bits(val))}
}

// Float32p constructs a field that carries a *float32. The returned Field will safely
// and explicitly represent `nil` when appropriate.
func Float32p(key string, val *float32) Field {
	if val == nil {
		return nilField(key)
	}
	return Float32(key, *val)
}

// Int constructs a field with the given key and value.
func Int(key string, val int) Field {
	return Int64(key, int64(val))
}

// Intp constructs a field that carries a *int. The returned Field will safely
// and explicitly represent `nil` when appropriate.
func Intp(key string, val *int) Field {
	if val == nil {
		return nilField(key)
	}
	return Int(key, *val)
}

// Int64 constructs a field with the given key and value.
func Int64(key string, val int64) Field {
	return Field{Key: key, Type: zapcore.Int64Type, Integer: val}
}

// Int64p constructs a field that carries a *int64. The returned Field will safely
// and explicitly represent `nil` when appropriate.
func Int64p(key string, val *int64) Field {
	if val == nil {
		return nilField(key)
	}
	return Int64(key, *val)
}

// Int32 constructs a field with the given key and value.
func Int32(key string, val int32) Field {
	return Field{Key: key, Type: zapcore.Int32Type, Integer: int64(val)}
}

// Int32p constructs a field that carries a *int32. The returned Field will safely
// and explicitly represent `nil` when appropriate.
func Int32p(key string, val *int32) Field {
	if val == nil {
		return nilField(key)
	}
	return Int32(key, *val)
}

// Int16 constructs a field with the given key and value.
func Int16(key string, val int16) Field {
	return Field{Key: key, Type: zapcore.Int16Type, Integer: int64(val)}
}

// Int16p constructs a field that carries a *int16. The returned Field will safely
// and explicitly represent `nil` when appropriate.
func Int16p(key string, val *int16) Field {
	if val == nil {
		return nilField(key)
	}
	return Int16(key, *val)
}

// Int8 constructs a field with the given key and value.
func Int8(key string, val int8) Field {
	return Field{Key: key, Type: zapcore.Int8Type, Integer: int64(val)}
}

// Int8p constructs a field that carries a *int8. The returned Field will safely
// and explicitly represent `nil` when appropriate.
func Int8p(key string, val *int8) Field {
	if val == nil {
		return nilField(key)
	}
	return Int8(key, *val)
}

// String constructs a field with the given key and value.
func String(key string, val string) Field {
	return Field{Key: key, Type: zapcore.StringType, String: val}
}

// Stringp constructs a field that carries a *string. The returned Field will safely
// and explicitly represent `nil` when appropriate.
func Stringp(key string, val *string) Field {
	if val == nil {
		return nilField(key)
	}
	return String(key, *val)
}

// Uint constructs a field with the given key and value.
func Uint(key string, val uint) Field {
	return Uint64(key, uint64(val))
}

// Uintp constructs a field that carries a *uint. The returned Field will safely
// and explicitly represent `nil` when appropriate.
func Uintp(key string, val *uint) Field {
	if val == nil {
		return nilField(key)
	}
	return Uint(key, *val)
}

// Uint64 constructs a field with the given key and value.
func Uint64(key string, val uint64) Field {
	return Field{Key: key, Type: zapcore.Uint64Type, Integer: int64(val)}
}

// Uint64p constructs a field that carries a *uint64. The returned Field will safely
// and explicitly represent `nil` when appropriate.
func Uint64p(key string, val *uint64) Field {
	if val == nil {
		return nilField(key)
	}
	return Uint64(key, *val)
}

// Uint32 constructs a field with the given key and value.
func Uint32(key string, val uint32) Field {
	return Field{Key: key, Type: zapcore.Uint32Type, Integer: int64(val)}
}

// Uint32p constructs a field that carries a *uint32. The returned Field will safely
// and explicitly represent `nil` when appropriate.
func Uint32p(key string, val *uint32) Field {
	if val == nil {
		return nilField(key)
	}
	return Uint32(key, *val)
}

// Uint16 constructs a field with the given key and value.
func Uint16(key string, val uint16) Field {
	return Field{Key: key, Type: zapcore.Uint16Type, Integer: int64(val)}
}

// Uint16p constructs a field that carries a *uint16. The returned Field will safely
// and explicitly represent `nil` when appropriate.
func Uint16p(key string, val *uint16) Field {
	if val == nil {
		return nilField(key)
	}
	return Uint16(key, *val)
}

// Uint8 constructs a field with the given key and value.
func Uint8(key string, val uint8) Field {
	return Field{Key: key, Type: zapcore.Uint8Type, Integer: int64(val)}
}

// Uint8p constructs a field that carries a *uint8. The returned Field will safely
// and explicitly represent `nil` when appropriate.
func Uint8p(key string, val *uint8) Field {
	if val == nil {
		return nilField(key)
	}
	return Uint8(key, *val)
}

// Uintptr constructs a field with the given key and value.
func Uintptr(key string, val uintptr) Field {
	return Field{Key: key, Type: zapcore.UintptrType, Integer: int64(val)}
}

// Uintptrp constructs a field that carries a *uintptr. The returned Field will safely
// and explicitly represent `nil` when appropriate.
func Uintptrp(key string, val *uintptr) Field {
	if val == nil {
		return nilField(key)
	}
	return Uintptr(key, *val)
}

// Reflect constructs a field with the given key and an arbitrary object. It uses
// an encoding-appropriate, reflection-based function to lazily serialize nearly
// any object into the logging context, but it's relatively slow and
// allocation-heavy. Outside tests, Any is always a better choice.
//
// If encoding fails (e.g., trying to serialize a map[int]string to JSON), Reflect
// includes the error message in the final log output.
func Reflect(key string, val interface{}) Field {
	return Field{Key: key, Type: zapcore.ReflectType, Interface: val}
}

// Namespace creates a named, isolated scope within the logger's context. All
// subsequent fields will be added to the new namespace.
//
// This helps prevent key collisions when injecting loggers into sub-components
// or third-party libraries.
func Namespace(key string) Field {
	return Field{Key: key, Type: zapcore.NamespaceType}
}

// Stringer constructs a field with the given key and the output of the value's
// String method. The Stringer's String method is called lazily.
func Stringer(key string, val fmt.Stringer) Field {
	return Field{Key: key, Type: zapcore.StringerType, Interface: val}
}

// Time constructs a Field with the given key and value. The encoder
// controls how the time is serialized.
func Time(key string, val time.Time) Field {
	if val.Before(_minTimeInt64) || val.After(_maxTimeInt64) {
		return Field{Key: key, Type: zapcore.TimeFullType, Interface: val}
	}
	return Field{Key: key, Type: zapcore.TimeType, Integer: val.UnixNano(), Interface: val.Location()}
}

// Timep constructs a field that carries a *time.Time. The returned Field will safely
// and explicitly represent `nil` when appropriate.
func Timep(key string, val *time.Time) Field {
	if val == nil {
		return nilField(key)
	}
	return Time(key, *val)
}

// Stack constructs a field that stores a stacktrace of the current goroutine
// under provided key. Keep in mind that taking a stacktrace is eager and
// expensive (relatively speaking); this function both makes an allocation and
// takes about two microseconds.
func Stack(key string) Field {
	return StackSkip(key, 1) // skip Stack
}

// StackSkip constructs a field similarly to Stack, but also skips the given
// number of frames from the top of the stacktrace.
func StackSkip(key string, skip int) Field {
	// Returning the stacktrace as a string costs an allocation, but saves us
	// from expanding the zapcore.Field union struct to include a byte slice. Since
	// taking a stacktrace is already so expensive (~10us), the extra allocation
	// is okay.
	return String(key, takeStacktrace(skip+1)) // skip StackSkip
}

// Duration constructs a field with the given key and value. The encoder
// controls how the duration is serialized.
func Duration(key string, val time.Duration) Field {
	return Field{Key: key, Type: zapcore.DurationType, Integer: int64(val)}
}

// Durationp constructs a field that carries a *time.Duration. The returned Field will safely
// and explicitly represent `nil` when appropriate.
func Durationp(key string, val *time.Duration) Field {
	if val == nil {
		return nilField(key)
	}
	return Duration(key, *val)
}

// Object constructs a field with the given key and ObjectMarshaler. It
// provides a flexible, but still type-safe and efficient, way to add map- or
// struct-like user-defined types to the logging context. The struct's
// MarshalLogObject method is called lazily.
func Object(key string, val zapcore.ObjectMarshaler) Field {
	return Field{Key: key, Type: zapcore.ObjectMarshalerType, Interface: val}
}

// Inline constructs a Field that is similar to Object, but it
// will add the elements of the provided ObjectMarshaler to the
// current namespace.
func Inline(val zapcore.ObjectMarshaler) Field {
	return zapcore.Field{
		Type:      zapcore.InlineMarshalerType,
		Interface: val,
	}
}

// Any takes a key and an arbitrary value and chooses the best way to represent
// them as a field, falling back to a reflection-based approach only if
// necessary.
//
// Since byte/uint8 and rune/int32 are aliases, Any can't differentiate between
// them. To minimize surprises, []byte values are treated as binary blobs, byte
// values are treated as uint8, and runes are always treated as integers.
func Any(key string, value interface{}) (f Field) {
	switch val := value.(type) {
	case zapcore.ObjectMarshaler:
		_object(unsafe.Pointer(&f), unsafe.Pointer(&key), unsafe.Pointer(&val))
	case zapcore.ArrayMarshaler:
		_array(unsafe.Pointer(&f), unsafe.Pointer(&key), unsafe.Pointer(&val))
	case bool:
		_bool(unsafe.Pointer(&f), unsafe.Pointer(&key), unsafe.Pointer(&val))
	case *bool:
		_boolp(unsafe.Pointer(&f), unsafe.Pointer(&key), unsafe.Pointer(&val))
	case []bool:
		_bools(unsafe.Pointer(&f), unsafe.Pointer(&key), unsafe.Pointer(&val))
	case complex128:
		_complex128(unsafe.Pointer(&f), unsafe.Pointer(&key), unsafe.Pointer(&val))
	case *complex128:
		_complex128p(unsafe.Pointer(&f), unsafe.Pointer(&key), unsafe.Pointer(&val))
	case []complex128:
		_complex128s(unsafe.Pointer(&f), unsafe.Pointer(&key), unsafe.Pointer(&val))
	case complex64:
		_complex64(unsafe.Pointer(&f), unsafe.Pointer(&key), unsafe.Pointer(&val))
	case *complex64:
		_complex64p(unsafe.Pointer(&f), unsafe.Pointer(&key), unsafe.Pointer(&val))
	case []complex64:
		_complex64s(unsafe.Pointer(&f), unsafe.Pointer(&key), unsafe.Pointer(&val))
	case float64:
		_float64(unsafe.Pointer(&f), unsafe.Pointer(&key), unsafe.Pointer(&val))
	case *float64:
		_float64p(unsafe.Pointer(&f), unsafe.Pointer(&key), unsafe.Pointer(&val))
	case []float64:
		_float64s(unsafe.Pointer(&f), unsafe.Pointer(&key), unsafe.Pointer(&val))
	case float32:
		_float32(unsafe.Pointer(&f), unsafe.Pointer(&key), unsafe.Pointer(&val))
	case *float32:
		_float32p(unsafe.Pointer(&f), unsafe.Pointer(&key), unsafe.Pointer(&val))
	case []float32:
		_float32s(unsafe.Pointer(&f), unsafe.Pointer(&key), unsafe.Pointer(&val))
	case int:
		_int(unsafe.Pointer(&f), unsafe.Pointer(&key), unsafe.Pointer(&val))
	case *int:
		_intp(unsafe.Pointer(&f), unsafe.Pointer(&key), unsafe.Pointer(&val))
	case []int:
		_ints(unsafe.Pointer(&f), unsafe.Pointer(&key), unsafe.Pointer(&val))
	case int64:
		_int64(unsafe.Pointer(&f), unsafe.Pointer(&key), unsafe.Pointer(&val))
	case *int64:
		_int64p(unsafe.Pointer(&f), unsafe.Pointer(&key), unsafe.Pointer(&val))
	case []int64:
		_int64s(unsafe.Pointer(&f), unsafe.Pointer(&key), unsafe.Pointer(&val))
	case int32:
		_int32(unsafe.Pointer(&f), unsafe.Pointer(&key), unsafe.Pointer(&val))
	case *int32:
		_int32p(unsafe.Pointer(&f), unsafe.Pointer(&key), unsafe.Pointer(&val))
	case []int32:
		_int32s(unsafe.Pointer(&f), unsafe.Pointer(&key), unsafe.Pointer(&val))
	case int16:
		_int16(unsafe.Pointer(&f), unsafe.Pointer(&key), unsafe.Pointer(&val))
	case *int16:
		_int16p(unsafe.Pointer(&f), unsafe.Pointer(&key), unsafe.Pointer(&val))
	case []int16:
		_int16s(unsafe.Pointer(&f), unsafe.Pointer(&key), unsafe.Pointer(&val))
	case int8:
		_int8(unsafe.Pointer(&f), unsafe.Pointer(&key), unsafe.Pointer(&val))
	case *int8:
		_int8p(unsafe.Pointer(&f), unsafe.Pointer(&key), unsafe.Pointer(&val))
	case []int8:
		_int8s(unsafe.Pointer(&f), unsafe.Pointer(&key), unsafe.Pointer(&val))
	case string:
		_string(unsafe.Pointer(&f), unsafe.Pointer(&key), unsafe.Pointer(&val))
	case *string:
		_stringp(unsafe.Pointer(&f), unsafe.Pointer(&key), unsafe.Pointer(&val))
	case []string:
		_strings(unsafe.Pointer(&f), unsafe.Pointer(&key), unsafe.Pointer(&val))
	case uint:
		_uint(unsafe.Pointer(&f), unsafe.Pointer(&key), unsafe.Pointer(&val))
	case *uint:
		_uintp(unsafe.Pointer(&f), unsafe.Pointer(&key), unsafe.Pointer(&val))
	case []uint:
		_uints(unsafe.Pointer(&f), unsafe.Pointer(&key), unsafe.Pointer(&val))
	case uint64:
		_uint64(unsafe.Pointer(&f), unsafe.Pointer(&key), unsafe.Pointer(&val))
	case *uint64:
		_uint64p(unsafe.Pointer(&f), unsafe.Pointer(&key), unsafe.Pointer(&val))
	case []uint64:
		_uint64s(unsafe.Pointer(&f), unsafe.Pointer(&key), unsafe.Pointer(&val))
	case uint32:
		_uint32(unsafe.Pointer(&f), unsafe.Pointer(&key), unsafe.Pointer(&val))
	case *uint32:
		_uint32p(unsafe.Pointer(&f), unsafe.Pointer(&key), unsafe.Pointer(&val))
	case []uint32:
		_uint32s(unsafe.Pointer(&f), unsafe.Pointer(&key), unsafe.Pointer(&val))
	case uint16:
		_uint16(unsafe.Pointer(&f), unsafe.Pointer(&key), unsafe.Pointer(&val))
	case *uint16:
		_uint16p(unsafe.Pointer(&f), unsafe.Pointer(&key), unsafe.Pointer(&val))
	case []uint16:
		_uint16s(unsafe.Pointer(&f), unsafe.Pointer(&key), unsafe.Pointer(&val))
	case uint8:
		_uint8(unsafe.Pointer(&f), unsafe.Pointer(&key), unsafe.Pointer(&val))
	case *uint8:
		_uint8p(unsafe.Pointer(&f), unsafe.Pointer(&key), unsafe.Pointer(&val))
	case []byte:
		_binary(unsafe.Pointer(&f), unsafe.Pointer(&key), unsafe.Pointer(&val))
	case uintptr:
		_uintptr(unsafe.Pointer(&f), unsafe.Pointer(&key), unsafe.Pointer(&val))
	case *uintptr:
		_uintptrp(unsafe.Pointer(&f), unsafe.Pointer(&key), unsafe.Pointer(&val))
	case []uintptr:
		_uintptrs(unsafe.Pointer(&f), unsafe.Pointer(&key), unsafe.Pointer(&val))
	case time.Time:
		_time(unsafe.Pointer(&f), unsafe.Pointer(&key), unsafe.Pointer(&val))
	case *time.Time:
		_timep(unsafe.Pointer(&f), unsafe.Pointer(&key), unsafe.Pointer(&val))
	case []time.Time:
		_times(unsafe.Pointer(&f), unsafe.Pointer(&key), unsafe.Pointer(&val))
	case time.Duration:
		_duration(unsafe.Pointer(&f), unsafe.Pointer(&key), unsafe.Pointer(&val))
	case *time.Duration:
		_durationp(unsafe.Pointer(&f), unsafe.Pointer(&key), unsafe.Pointer(&val))
	case []time.Duration:
		_durations(unsafe.Pointer(&f), unsafe.Pointer(&key), unsafe.Pointer(&val))
	case error:
		_namedError(unsafe.Pointer(&f), unsafe.Pointer(&key), unsafe.Pointer(&val))
	case []error:
		_errors(unsafe.Pointer(&f), unsafe.Pointer(&key), unsafe.Pointer(&val))
	case fmt.Stringer:
		_stringer(unsafe.Pointer(&f), unsafe.Pointer(&key), unsafe.Pointer(&val))
	default:
		_reflect(unsafe.Pointer(&f), unsafe.Pointer(&key), unsafe.Pointer(&val))
	}

	return
}

//go:noinline
//go:nosplit
func _object(f unsafe.Pointer, k unsafe.Pointer, value unsafe.Pointer) {
	*((*Field)(f)) = Object(*(*string)(k), *(*zapcore.ObjectMarshaler)(value))
}

//go:noinline
//go:nosplit
func _array(f unsafe.Pointer, k unsafe.Pointer, value unsafe.Pointer) {
	*((*Field)(f)) = Array(*(*string)(k), *(*zapcore.ArrayMarshaler)(value))
}

//go:noinline
//go:nosplit
func _bool(f unsafe.Pointer, k unsafe.Pointer, value unsafe.Pointer) {
	*((*Field)(f)) = Bool(*(*string)(k), *(*bool)(value))
}

//go:noinline
//go:nosplit
func _boolp(f unsafe.Pointer, k unsafe.Pointer, value unsafe.Pointer) {
	*((*Field)(f)) = Boolp(*(*string)(k), *(**bool)(value))
}

//go:noinline
//go:nosplit
func _bools(f unsafe.Pointer, k unsafe.Pointer, value unsafe.Pointer) {
	*((*Field)(f)) = Bools(*(*string)(k), *(*[]bool)(value))
}

//go:noinline
//go:nosplit
func _complex128(f unsafe.Pointer, k unsafe.Pointer, value unsafe.Pointer) {
	*((*Field)(f)) = Complex128(*(*string)(k), *(*complex128)(value))
}

//go:noinline
//go:nosplit
func _complex128p(f unsafe.Pointer, k unsafe.Pointer, value unsafe.Pointer) {
	*((*Field)(f)) = Complex128p(*(*string)(k), *(**complex128)(value))
}

//go:noinline
//go:nosplit
func _complex128s(f unsafe.Pointer, k unsafe.Pointer, value unsafe.Pointer) {
	*((*Field)(f)) = Complex128s(*(*string)(k), *(*[]complex128)(value))
}

//go:noinline
//go:nosplit
func _complex64(f unsafe.Pointer, k unsafe.Pointer, value unsafe.Pointer) {
	*((*Field)(f)) = Complex64(*(*string)(k), *(*complex64)(value))
}

//go:noinline
//go:nosplit
func _complex64p(f unsafe.Pointer, k unsafe.Pointer, value unsafe.Pointer) {
	*((*Field)(f)) = Complex64p(*(*string)(k), *(**complex64)(value))
}

//go:noinline
//go:nosplit
func _complex64s(f unsafe.Pointer, k unsafe.Pointer, value unsafe.Pointer) {
	*((*Field)(f)) = Complex64s(*(*string)(k), *(*[]complex64)(value))
}

//go:noinline
//go:nosplit
func _float64(f unsafe.Pointer, k unsafe.Pointer, value unsafe.Pointer) {
	*((*Field)(f)) = Float64(*(*string)(k), *(*float64)(value))
}

//go:noinline
//go:nosplit
func _float64p(f unsafe.Pointer, k unsafe.Pointer, value unsafe.Pointer) {
	*((*Field)(f)) = Float64p(*(*string)(k), *(**float64)(value))
}

//go:noinline
//go:nosplit
func _float64s(f unsafe.Pointer, k unsafe.Pointer, value unsafe.Pointer) {
	*((*Field)(f)) = Float64s(*(*string)(k), *(*[]float64)(value))
}

//go:noinline
//go:nosplit
func _float32(f unsafe.Pointer, k unsafe.Pointer, value unsafe.Pointer) {
	*((*Field)(f)) = Float32(*(*string)(k), *(*float32)(value))
}

//go:noinline
//go:nosplit
func _float32p(f unsafe.Pointer, k unsafe.Pointer, value unsafe.Pointer) {
	*((*Field)(f)) = Float32p(*(*string)(k), *(**float32)(value))
}

//go:noinline
//go:nosplit
func _float32s(f unsafe.Pointer, k unsafe.Pointer, value unsafe.Pointer) {
	*((*Field)(f)) = Float32s(*(*string)(k), *(*[]float32)(value))
}

//go:noinline
//go:nosplit
func _int(f unsafe.Pointer, k unsafe.Pointer, value unsafe.Pointer) {
	*((*Field)(f)) = Int(*(*string)(k), *(*int)(value))
}

//go:noinline
//go:nosplit
func _intp(f unsafe.Pointer, k unsafe.Pointer, value unsafe.Pointer) {
	*((*Field)(f)) = Intp(*(*string)(k), *(**int)(value))
}

//go:noinline
//go:nosplit
func _ints(f unsafe.Pointer, k unsafe.Pointer, value unsafe.Pointer) {
	*((*Field)(f)) = Ints(*(*string)(k), *(*[]int)(value))
}

//go:noinline
//go:nosplit
func _int64(f unsafe.Pointer, k unsafe.Pointer, value unsafe.Pointer) {
	*((*Field)(f)) = Int64(*(*string)(k), *(*int64)(value))
}

//go:noinline
//go:nosplit
func _int64p(f unsafe.Pointer, k unsafe.Pointer, value unsafe.Pointer) {
	*((*Field)(f)) = Int64p(*(*string)(k), *(**int64)(value))
}

//go:noinline
//go:nosplit
func _int64s(f unsafe.Pointer, k unsafe.Pointer, value unsafe.Pointer) {
	*((*Field)(f)) = Int64s(*(*string)(k), *(*[]int64)(value))
}

//go:noinline
//go:nosplit
func _int32(f unsafe.Pointer, k unsafe.Pointer, value unsafe.Pointer) {
	*((*Field)(f)) = Int32(*(*string)(k), *(*int32)(value))
}

//go:noinline
//go:nosplit
func _int32p(f unsafe.Pointer, k unsafe.Pointer, value unsafe.Pointer) {
	*((*Field)(f)) = Int32p(*(*string)(k), *(**int32)(value))
}

//go:noinline
//go:nosplit
func _int32s(f unsafe.Pointer, k unsafe.Pointer, value unsafe.Pointer) {
	*((*Field)(f)) = Int32s(*(*string)(k), *(*[]int32)(value))
}

//go:noinline
//go:nosplit
func _int16(f unsafe.Pointer, k unsafe.Pointer, value unsafe.Pointer) {
	*((*Field)(f)) = Int16(*(*string)(k), *(*int16)(value))
}

//go:noinline
//go:nosplit
func _int16p(f unsafe.Pointer, k unsafe.Pointer, value unsafe.Pointer) {
	*((*Field)(f)) = Int16p(*(*string)(k), *(**int16)(value))
}

//go:noinline
//go:nosplit
func _int16s(f unsafe.Pointer, k unsafe.Pointer, value unsafe.Pointer) {
	*((*Field)(f)) = Int16s(*(*string)(k), *(*[]int16)(value))
}

//go:noinline
//go:nosplit
func _int8(f unsafe.Pointer, k unsafe.Pointer, value unsafe.Pointer) {
	*((*Field)(f)) = Int8(*(*string)(k), *(*int8)(value))
}

//go:noinline
//go:nosplit
func _int8p(f unsafe.Pointer, k unsafe.Pointer, value unsafe.Pointer) {
	*((*Field)(f)) = Int8p(*(*string)(k), *(**int8)(value))
}

//go:noinline
//go:nosplit
func _int8s(f unsafe.Pointer, k unsafe.Pointer, value unsafe.Pointer) {
	*((*Field)(f)) = Int8s(*(*string)(k), *(*[]int8)(value))
}

//go:noinline
//go:nosplit
func _string(f unsafe.Pointer, k unsafe.Pointer, value unsafe.Pointer) {
	*((*Field)(f)) = String(*(*string)(k), *(*string)(value))
}

//go:noinline
//go:nosplit
func _stringp(f unsafe.Pointer, k unsafe.Pointer, value unsafe.Pointer) {
	*((*Field)(f)) = Stringp(*(*string)(k), *(**string)(value))
}

//go:noinline
//go:nosplit
func _strings(f unsafe.Pointer, k unsafe.Pointer, value unsafe.Pointer) {
	*((*Field)(f)) = Strings(*(*string)(k), *(*[]string)(value))
}

//go:noinline
//go:nosplit
func _uint(f unsafe.Pointer, k unsafe.Pointer, value unsafe.Pointer) {
	*((*Field)(f)) = Uint(*(*string)(k), *(*uint)(value))
}

//go:noinline
//go:nosplit
func _uintp(f unsafe.Pointer, k unsafe.Pointer, value unsafe.Pointer) {
	*((*Field)(f)) = Uintp(*(*string)(k), *(**uint)(value))
}

//go:noinline
//go:nosplit
func _uints(f unsafe.Pointer, k unsafe.Pointer, value unsafe.Pointer) {
	*((*Field)(f)) = Uints(*(*string)(k), *(*[]uint)(value))
}

//go:noinline
//go:nosplit
func _uint64(f unsafe.Pointer, k unsafe.Pointer, value unsafe.Pointer) {
	*((*Field)(f)) = Uint64(*(*string)(k), *(*uint64)(value))
}

//go:noinline
//go:nosplit
func _uint64p(f unsafe.Pointer, k unsafe.Pointer, value unsafe.Pointer) {
	*((*Field)(f)) = Uint64p(*(*string)(k), *(**uint64)(value))
}

//go:noinline
//go:nosplit
func _uint64s(f unsafe.Pointer, k unsafe.Pointer, value unsafe.Pointer) {
	*((*Field)(f)) = Uint64s(*(*string)(k), *(*[]uint64)(value))
}

//go:noinline
//go:nosplit
func _uint32(f unsafe.Pointer, k unsafe.Pointer, value unsafe.Pointer) {
	*((*Field)(f)) = Uint32(*(*string)(k), *(*uint32)(value))
}

//go:noinline
//go:nosplit
func _uint32p(f unsafe.Pointer, k unsafe.Pointer, value unsafe.Pointer) {
	*((*Field)(f)) = Uint32p(*(*string)(k), *(**uint32)(value))
}

//go:noinline
//go:nosplit
func _uint32s(f unsafe.Pointer, k unsafe.Pointer, value unsafe.Pointer) {
	*((*Field)(f)) = Uint32s(*(*string)(k), *(*[]uint32)(value))
}

//go:noinline
//go:nosplit
func _uint16(f unsafe.Pointer, k unsafe.Pointer, value unsafe.Pointer) {
	*((*Field)(f)) = Uint16(*(*string)(k), *(*uint16)(value))
}

//go:noinline
//go:nosplit
func _uint16p(f unsafe.Pointer, k unsafe.Pointer, value unsafe.Pointer) {
	*((*Field)(f)) = Uint16p(*(*string)(k), *(**uint16)(value))
}

//go:noinline
//go:nosplit
func _uint16s(f unsafe.Pointer, k unsafe.Pointer, value unsafe.Pointer) {
	*((*Field)(f)) = Uint16s(*(*string)(k), *(*[]uint16)(value))
}

//go:noinline
//go:nosplit
func _uint8(f unsafe.Pointer, k unsafe.Pointer, value unsafe.Pointer) {
	*((*Field)(f)) = Uint8(*(*string)(k), *(*uint8)(value))
}

//go:noinline
//go:nosplit
func _uint8p(f unsafe.Pointer, k unsafe.Pointer, value unsafe.Pointer) {
	*((*Field)(f)) = Uint8p(*(*string)(k), *(**uint8)(value))
}

//go:noinline
//go:nosplit
func _binary(f unsafe.Pointer, k unsafe.Pointer, value unsafe.Pointer) {
	*((*Field)(f)) = Binary(*(*string)(k), *(*[]byte)(value))
}

//go:noinline
//go:nosplit
func _uintptr(f unsafe.Pointer, k unsafe.Pointer, value unsafe.Pointer) {
	*((*Field)(f)) = Uintptr(*(*string)(k), *(*uintptr)(value))
}

//go:noinline
//go:nosplit
func _uintptrp(f unsafe.Pointer, k unsafe.Pointer, value unsafe.Pointer) {
	*((*Field)(f)) = Uintptrp(*(*string)(k), *(**uintptr)(value))
}

//go:noinline
//go:nosplit
func _uintptrs(f unsafe.Pointer, k unsafe.Pointer, value unsafe.Pointer) {
	*((*Field)(f)) = Uintptrs(*(*string)(k), *(*[]uintptr)(value))
}

//go:noinline
//go:nosplit
func _time(f unsafe.Pointer, k unsafe.Pointer, value unsafe.Pointer) {
	*((*Field)(f)) = Time(*(*string)(k), *(*time.Time)(value))
}

//go:noinline
//go:nosplit
func _timep(f unsafe.Pointer, k unsafe.Pointer, value unsafe.Pointer) {
	*((*Field)(f)) = Timep(*(*string)(k), *(**time.Time)(value))
}

//go:noinline
//go:nosplit
func _times(f unsafe.Pointer, k unsafe.Pointer, value unsafe.Pointer) {
	*((*Field)(f)) = Times(*(*string)(k), *(*[]time.Time)(value))
}

//go:noinline
//go:nosplit
func _duration(f unsafe.Pointer, k unsafe.Pointer, value unsafe.Pointer) {
	*((*Field)(f)) = Duration(*(*string)(k), *(*time.Duration)(value))
}

//go:noinline
//go:nosplit
func _durationp(f unsafe.Pointer, k unsafe.Pointer, value unsafe.Pointer) {
	*((*Field)(f)) = Durationp(*(*string)(k), *(**time.Duration)(value))
}

//go:noinline
//go:nosplit
func _durations(f unsafe.Pointer, k unsafe.Pointer, value unsafe.Pointer) {
	*((*Field)(f)) = Durations(*(*string)(k), *(*[]time.Duration)(value))
}

//go:noinline
//go:nosplit
func _namedError(f unsafe.Pointer, k unsafe.Pointer, value unsafe.Pointer) {
	*((*Field)(f)) = NamedError(*(*string)(k), *(*error)(value))
}

//go:noinline
//go:nosplit
func _errors(f unsafe.Pointer, k unsafe.Pointer, value unsafe.Pointer) {
	*((*Field)(f)) = Errors(*(*string)(k), *(*[]error)(value))
}

//go:noinline
func _stringer(f unsafe.Pointer, k unsafe.Pointer, value unsafe.Pointer) {
	*((*Field)(f)) = Stringer(*(*string)(k), *(*fmt.Stringer)(value))
}

//go:noinline
func _reflect(f unsafe.Pointer, k unsafe.Pointer, value unsafe.Pointer) {
	*((*Field)(f)) = Reflect(*(*string)(k), *(*any)(value))
}
