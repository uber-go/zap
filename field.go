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
	"reflect"
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
	// this can hold the memory
	storage := make([]byte, _minimumSize)
	copy(storage[_sizeOfField:], (*(*[_sizeOfString]byte)(unsafe.Pointer(&key)))[:])
	copy(storage[len(storage)-2*_sizeOfPtr:], (*(*[2 * _sizeOfPtr]byte)(unsafe.Pointer(&value)))[:])

	switch val := value.(type) {
	case zapcore.ObjectMarshaler:
		// fix interface type in temporal storage
		copy(storage[len(storage)-2*_sizeOfPtr:], (*(*[2 * _sizeOfPtr]byte)(unsafe.Pointer(&val)))[:])

		_object(&storage)
	case zapcore.ArrayMarshaler:
		// fix interface type in temporal storage
		copy(storage[len(storage)-2*_sizeOfPtr:], (*(*[2 * _sizeOfPtr]byte)(unsafe.Pointer(&val)))[:])

		_array(&storage)
	case bool:
		_bool(&storage)
	case *bool:
		_boolp(&storage)
	case []bool:
		_bools(&storage)
	case complex128:
		_complex128(&storage)
	case *complex128:
		_complex128p(&storage)
	case []complex128:
		_complex128s(&storage)
	case complex64:
		_complex64(&storage)
	case *complex64:
		_complex64p(&storage)
	case []complex64:
		_complex64s(&storage)
	case float64:
		_float64(&storage)
	case *float64:
		_float64p(&storage)
	case []float64:
		_float64s(&storage)
	case float32:
		_float32(&storage)
	case *float32:
		_float32p(&storage)
	case []float32:
		_float32s(&storage)
	case int:
		_int(&storage)
	case *int:
		_intp(&storage)
	case []int:
		_ints(&storage)
	case int64:
		_int64(&storage)
	case *int64:
		_int64p(&storage)
	case []int64:
		_int64s(&storage)
	case int32:
		_int32(&storage)
	case *int32:
		_int32p(&storage)
	case []int32:
		_int32s(&storage)
	case int16:
		_int16(&storage)
	case *int16:
		_int16p(&storage)
	case []int16:
		_int16s(&storage)
	case int8:
		_int8(&storage)
	case *int8:
		_int8p(&storage)
	case []int8:
		_int8s(&storage)
	case string:
		_string(&storage)
	case *string:
		_stringp(&storage)
	case []string:
		_strings(&storage)
	case uint:
		_uint(&storage)
	case *uint:
		_uintp(&storage)
	case []uint:
		_uints(&storage)
	case uint64:
		_uint64(&storage)
	case *uint64:
		_uint64p(&storage)
	case []uint64:
		_uint64s(&storage)
	case uint32:
		_uint32(&storage)
	case *uint32:
		_uint32p(&storage)
	case []uint32:
		_uint32s(&storage)
	case uint16:
		_uint16(&storage)
	case *uint16:
		_uint16p(&storage)
	case []uint16:
		_uint16s(&storage)
	case uint8:
		_uint8(&storage)
	case *uint8:
		_uint8p(&storage)
	case []byte:
		_binary(&storage)
	case uintptr:
		_uintptr(&storage)
	case *uintptr:
		_uintptrp(&storage)
	case []uintptr:
		_uintptrs(&storage)
	case time.Time:
		_time(&storage)
	case *time.Time:
		_timep(&storage)
	case []time.Time:
		_times(&storage)
	case time.Duration:
		_duration(&storage)
	case *time.Duration:
		_durationp(&storage)
	case []time.Duration:
		_durations(&storage)
	case error:
		// fix interface type in temporal storage
		copy(storage[len(storage)-2*_sizeOfPtr:], (*(*[2 * _sizeOfPtr]byte)(unsafe.Pointer(&val)))[:])

		_namedError(&storage)
	case []error:
		_errors(&storage)
	case fmt.Stringer:
		// fix interface type in temporal storage
		copy(storage[len(storage)-2*_sizeOfPtr:], (*(*[2 * _sizeOfPtr]byte)(unsafe.Pointer(&val)))[:])

		_stringer(&storage)
	default:
		_reflect(&storage)
	}

	f = *(*Field)(unsafe.Pointer((*sliceHeader)(unsafe.Pointer(&storage)).Data))

	return
}

//go:noinline
//go:nosplit
func _object(p *[]byte) {
	ptr := unsafe.Pointer((*sliceHeader)(unsafe.Pointer(p)).Data)
	// zapcore.ObjectMarshaler is an interface so there is no need for temporal usafe.Pointer
	*((*Field)(ptr)) = Object(*(*string)(unsafe.Add(ptr, _sizeOfField)),
		*(*zapcore.ObjectMarshaler)(unsafe.Add(ptr, _sizeOfField+_sizeOfString)))
}

//go:noinline
//go:nosplit
func _array(p *[]byte) {
	ptr := unsafe.Pointer((*sliceHeader)(unsafe.Pointer(p)).Data)
	// zapcore.ArrayMarshaler is an interface so there is no need for temporal usafe.Pointer
	*((*Field)(ptr)) = Array(*(*string)(unsafe.Add(ptr, _sizeOfField)),
		*(*zapcore.ArrayMarshaler)(unsafe.Add(ptr, _sizeOfField+_sizeOfString)))
}

//go:noinline
//go:nosplit
func _bool(p *[]byte) {
	ptr := unsafe.Pointer((*sliceHeader)(unsafe.Pointer(p)).Data)
	*((*Field)(ptr)) = Bool(*(*string)(unsafe.Add(ptr, _sizeOfField)),
		*(*bool)(*(*unsafe.Pointer)(unsafe.Add(ptr, _sizeOfField+_sizeOfString+_sizeOfPtr))))
}

//go:noinline
//go:nosplit
func _boolp(p *[]byte) {
	ptr := unsafe.Pointer((*sliceHeader)(unsafe.Pointer(p)).Data)
	*((*Field)(ptr)) = Boolp(*(*string)(unsafe.Add(ptr, _sizeOfField)),
		(*bool)(*(*unsafe.Pointer)(unsafe.Add(ptr, _sizeOfField+_sizeOfString+_sizeOfPtr))))
}

//go:noinline
//go:nosplit
func _bools(p *[]byte) {
	ptr := unsafe.Pointer((*sliceHeader)(unsafe.Pointer(p)).Data)
	*((*Field)(ptr)) = Bools(*(*string)(unsafe.Add(ptr, _sizeOfField)),
		*(*[]bool)(*(*unsafe.Pointer)(unsafe.Add(ptr, _sizeOfField+_sizeOfString+_sizeOfPtr))))
}

//go:noinline
//go:nosplit
func _complex128(p *[]byte) {
	ptr := unsafe.Pointer((*sliceHeader)(unsafe.Pointer(p)).Data)
	*((*Field)(ptr)) = Complex128(*(*string)(unsafe.Add(ptr, _sizeOfField)),
		*(*complex128)(*(*unsafe.Pointer)(unsafe.Add(ptr, _sizeOfField+_sizeOfString+_sizeOfPtr))))
}

//go:noinline
//go:nosplit
func _complex128p(p *[]byte) {
	ptr := unsafe.Pointer((*sliceHeader)(unsafe.Pointer(p)).Data)
	*((*Field)(ptr)) = Complex128p(*(*string)(unsafe.Add(ptr, _sizeOfField)),
		(*complex128)(*(*unsafe.Pointer)(unsafe.Add(ptr, _sizeOfField+_sizeOfString+_sizeOfPtr))))
}

//go:noinline
//go:nosplit
func _complex128s(p *[]byte) {
	ptr := unsafe.Pointer((*sliceHeader)(unsafe.Pointer(p)).Data)
	*((*Field)(ptr)) = Complex128s(*(*string)(unsafe.Add(ptr, _sizeOfField)),
		*(*[]complex128)(*(*unsafe.Pointer)(unsafe.Add(ptr, _sizeOfField+_sizeOfString+_sizeOfPtr))))
}

//go:noinline
//go:nosplit
func _complex64(p *[]byte) {
	ptr := unsafe.Pointer((*sliceHeader)(unsafe.Pointer(p)).Data)
	*((*Field)(ptr)) = Complex64(*(*string)(unsafe.Add(ptr, _sizeOfField)),
		*(*complex64)(*(*unsafe.Pointer)(unsafe.Add(ptr, _sizeOfField+_sizeOfString+_sizeOfPtr))))
}

//go:noinline
//go:nosplit
func _complex64p(p *[]byte) {
	ptr := unsafe.Pointer((*sliceHeader)(unsafe.Pointer(p)).Data)
	*((*Field)(ptr)) = Complex64p(*(*string)(unsafe.Add(ptr, _sizeOfField)),
		(*complex64)(*(*unsafe.Pointer)(unsafe.Add(ptr, _sizeOfField+_sizeOfString+_sizeOfPtr))))
}

//go:noinline
//go:nosplit
func _complex64s(p *[]byte) {
	ptr := unsafe.Pointer((*sliceHeader)(unsafe.Pointer(p)).Data)
	*((*Field)(ptr)) = Complex64s(*(*string)(unsafe.Add(ptr, _sizeOfField)),
		*(*[]complex64)(*(*unsafe.Pointer)(unsafe.Add(ptr, _sizeOfField+_sizeOfString+_sizeOfPtr))))
}

//go:noinline
//go:nosplit
func _float64(p *[]byte) {
	ptr := unsafe.Pointer((*sliceHeader)(unsafe.Pointer(p)).Data)
	*((*Field)(ptr)) = Float64(*(*string)(unsafe.Add(ptr, _sizeOfField)),
		*(*float64)(*(*unsafe.Pointer)(unsafe.Add(ptr, _sizeOfField+_sizeOfString+_sizeOfPtr))))
}

//go:noinline
//go:nosplit
func _float64p(p *[]byte) {
	ptr := unsafe.Pointer((*sliceHeader)(unsafe.Pointer(p)).Data)
	*((*Field)(ptr)) = Float64p(*(*string)(unsafe.Add(ptr, _sizeOfField)),
		(*float64)(*(*unsafe.Pointer)(unsafe.Add(ptr, _sizeOfField+_sizeOfString+_sizeOfPtr))))
}

//go:noinline
//go:nosplit
func _float64s(p *[]byte) {
	ptr := unsafe.Pointer((*sliceHeader)(unsafe.Pointer(p)).Data)
	*((*Field)(ptr)) = Float64s(*(*string)(unsafe.Add(ptr, _sizeOfField)),
		*(*[]float64)(*(*unsafe.Pointer)(unsafe.Add(ptr, _sizeOfField+_sizeOfString+_sizeOfPtr))))
}

//go:noinline
//go:nosplit
func _float32(p *[]byte) {
	ptr := unsafe.Pointer((*sliceHeader)(unsafe.Pointer(p)).Data)
	*((*Field)(ptr)) = Float32(*(*string)(unsafe.Add(ptr, _sizeOfField)),
		*(*float32)(*(*unsafe.Pointer)(unsafe.Add(ptr, _sizeOfField+_sizeOfString+_sizeOfPtr))))
}

//go:noinline
//go:nosplit
func _float32p(p *[]byte) {
	ptr := unsafe.Pointer((*sliceHeader)(unsafe.Pointer(p)).Data)
	*((*Field)(ptr)) = Float32p(*(*string)(unsafe.Add(ptr, _sizeOfField)),
		(*float32)(*(*unsafe.Pointer)(unsafe.Add(ptr, _sizeOfField+_sizeOfString+_sizeOfPtr))))
}

//go:noinline
//go:nosplit
func _float32s(p *[]byte) {
	ptr := unsafe.Pointer((*sliceHeader)(unsafe.Pointer(p)).Data)
	*((*Field)(ptr)) = Float32s(*(*string)(unsafe.Add(ptr, _sizeOfField)),
		*(*[]float32)(*(*unsafe.Pointer)(unsafe.Add(ptr, _sizeOfField+_sizeOfString+_sizeOfPtr))))
}

//go:noinline
//go:nosplit
func _int(p *[]byte) {
	ptr := unsafe.Pointer((*sliceHeader)(unsafe.Pointer(p)).Data)
	*((*Field)(ptr)) = Int(*(*string)(unsafe.Add(ptr, _sizeOfField)),
		*(*int)(*(*unsafe.Pointer)(unsafe.Add(ptr, _sizeOfField+_sizeOfString+_sizeOfPtr))))
}

//go:noinline
//go:nosplit
func _intp(p *[]byte) {
	ptr := unsafe.Pointer((*sliceHeader)(unsafe.Pointer(p)).Data)
	*((*Field)(ptr)) = Intp(*(*string)(unsafe.Add(ptr, _sizeOfField)),
		(*int)(*(*unsafe.Pointer)(unsafe.Add(ptr, _sizeOfField+_sizeOfString+_sizeOfPtr))))
}

//go:noinline
//go:nosplit
func _ints(p *[]byte) {
	ptr := unsafe.Pointer((*sliceHeader)(unsafe.Pointer(p)).Data)
	*((*Field)(ptr)) = Ints(*(*string)(unsafe.Add(ptr, _sizeOfField)),
		*(*[]int)(*(*unsafe.Pointer)(unsafe.Add(ptr, _sizeOfField+_sizeOfString+_sizeOfPtr))))
}

//go:noinline
//go:nosplit
func _int64(p *[]byte) {
	ptr := unsafe.Pointer((*sliceHeader)(unsafe.Pointer(p)).Data)
	*((*Field)(ptr)) = Int64(*(*string)(unsafe.Add(ptr, _sizeOfField)),
		*(*int64)(*(*unsafe.Pointer)(unsafe.Add(ptr, _sizeOfField+_sizeOfString+_sizeOfPtr))))
}

//go:noinline
//go:nosplit
func _int64p(p *[]byte) {
	ptr := unsafe.Pointer((*sliceHeader)(unsafe.Pointer(p)).Data)
	*((*Field)(ptr)) = Int64p(*(*string)(unsafe.Add(ptr, _sizeOfField)),
		(*int64)(*(*unsafe.Pointer)(unsafe.Add(ptr, _sizeOfField+_sizeOfString+_sizeOfPtr))))
}

//go:noinline
//go:nosplit
func _int64s(p *[]byte) {
	ptr := unsafe.Pointer((*sliceHeader)(unsafe.Pointer(p)).Data)
	*((*Field)(ptr)) = Int64s(*(*string)(unsafe.Add(ptr, _sizeOfField)),
		*(*[]int64)(*(*unsafe.Pointer)(unsafe.Add(ptr, _sizeOfField+_sizeOfString+_sizeOfPtr))))
}

//go:noinline
//go:nosplit
func _int32(p *[]byte) {
	ptr := unsafe.Pointer((*sliceHeader)(unsafe.Pointer(p)).Data)
	*((*Field)(ptr)) = Int32(*(*string)(unsafe.Add(ptr, _sizeOfField)),
		*(*int32)(*(*unsafe.Pointer)(unsafe.Add(ptr, _sizeOfField+_sizeOfString+_sizeOfPtr))))
}

//go:noinline
//go:nosplit
func _int32p(p *[]byte) {
	ptr := unsafe.Pointer((*sliceHeader)(unsafe.Pointer(p)).Data)
	*((*Field)(ptr)) = Int32p(*(*string)(unsafe.Add(ptr, _sizeOfField)),
		(*int32)(*(*unsafe.Pointer)(unsafe.Add(ptr, _sizeOfField+_sizeOfString+_sizeOfPtr))))
}

//go:noinline
//go:nosplit
func _int32s(p *[]byte) {
	ptr := unsafe.Pointer((*sliceHeader)(unsafe.Pointer(p)).Data)
	*((*Field)(ptr)) = Int32s(*(*string)(unsafe.Add(ptr, _sizeOfField)),
		*(*[]int32)(*(*unsafe.Pointer)(unsafe.Add(ptr, _sizeOfField+_sizeOfString+_sizeOfPtr))))
}

//go:noinline
//go:nosplit
func _int16(p *[]byte) {
	ptr := unsafe.Pointer((*sliceHeader)(unsafe.Pointer(p)).Data)
	*((*Field)(ptr)) = Int16(*(*string)(unsafe.Add(ptr, _sizeOfField)),
		*(*int16)(*(*unsafe.Pointer)(unsafe.Add(ptr, _sizeOfField+_sizeOfString+_sizeOfPtr))))
}

//go:noinline
//go:nosplit
func _int16p(p *[]byte) {
	ptr := unsafe.Pointer((*sliceHeader)(unsafe.Pointer(p)).Data)
	*((*Field)(ptr)) = Int16p(*(*string)(unsafe.Add(ptr, _sizeOfField)),
		(*int16)(*(*unsafe.Pointer)(unsafe.Add(ptr, _sizeOfField+_sizeOfString+_sizeOfPtr))))
}

//go:noinline
//go:nosplit
func _int16s(p *[]byte) {
	ptr := unsafe.Pointer((*sliceHeader)(unsafe.Pointer(p)).Data)
	*((*Field)(ptr)) = Int16s(*(*string)(unsafe.Add(ptr, _sizeOfField)),
		*(*[]int16)(*(*unsafe.Pointer)(unsafe.Add(ptr, _sizeOfField+_sizeOfString+_sizeOfPtr))))
}

//go:noinline
//go:nosplit
func _int8(p *[]byte) {
	ptr := unsafe.Pointer((*sliceHeader)(unsafe.Pointer(p)).Data)
	*((*Field)(ptr)) = Int8(*(*string)(unsafe.Add(ptr, _sizeOfField)),
		*(*int8)(*(*unsafe.Pointer)(unsafe.Add(ptr, _sizeOfField+_sizeOfString+_sizeOfPtr))))
}

//go:noinline
//go:nosplit
func _int8p(p *[]byte) {
	ptr := unsafe.Pointer((*sliceHeader)(unsafe.Pointer(p)).Data)
	*((*Field)(ptr)) = Int8p(*(*string)(unsafe.Add(ptr, _sizeOfField)),
		(*int8)(*(*unsafe.Pointer)(unsafe.Add(ptr, _sizeOfField+_sizeOfString+_sizeOfPtr))))
}

//go:noinline
//go:nosplit
func _int8s(p *[]byte) {
	ptr := unsafe.Pointer((*sliceHeader)(unsafe.Pointer(p)).Data)
	*((*Field)(ptr)) = Int8s(*(*string)(unsafe.Add(ptr, _sizeOfField)),
		*(*[]int8)(*(*unsafe.Pointer)(unsafe.Add(ptr, _sizeOfField+_sizeOfString+_sizeOfPtr))))
}

//go:noinline
//go:nosplit
func _string(p *[]byte) {
	ptr := unsafe.Pointer((*sliceHeader)(unsafe.Pointer(p)).Data)
	*((*Field)(ptr)) = String(*(*string)(unsafe.Add(ptr, _sizeOfField)),
		*(*string)(*(*unsafe.Pointer)(unsafe.Add(ptr, _sizeOfField+_sizeOfString+_sizeOfPtr))))
}

//go:noinline
//go:nosplit
func _stringp(p *[]byte) {
	ptr := unsafe.Pointer((*sliceHeader)(unsafe.Pointer(p)).Data)
	*((*Field)(ptr)) = Stringp(*(*string)(unsafe.Add(ptr, _sizeOfField)),
		(*string)(*(*unsafe.Pointer)(unsafe.Add(ptr, _sizeOfField+_sizeOfString+_sizeOfPtr))))
}

//go:noinline
//go:nosplit
func _strings(p *[]byte) {
	ptr := unsafe.Pointer((*sliceHeader)(unsafe.Pointer(p)).Data)
	*((*Field)(ptr)) = Strings(*(*string)(unsafe.Add(ptr, _sizeOfField)),
		*(*[]string)(*(*unsafe.Pointer)(unsafe.Add(ptr, _sizeOfField+_sizeOfString+_sizeOfPtr))))
}

//go:noinline
//go:nosplit
func _uint(p *[]byte) {
	ptr := unsafe.Pointer((*sliceHeader)(unsafe.Pointer(p)).Data)
	*((*Field)(ptr)) = Uint(*(*string)(unsafe.Add(ptr, _sizeOfField)),
		*(*uint)(*(*unsafe.Pointer)(unsafe.Add(ptr, _sizeOfField+_sizeOfString+_sizeOfPtr))))
}

//go:noinline
//go:nosplit
func _uintp(p *[]byte) {
	ptr := unsafe.Pointer((*sliceHeader)(unsafe.Pointer(p)).Data)
	*((*Field)(ptr)) = Uintp(*(*string)(unsafe.Add(ptr, _sizeOfField)),
		(*uint)(*(*unsafe.Pointer)(unsafe.Add(ptr, _sizeOfField+_sizeOfString+_sizeOfPtr))))
}

//go:noinline
//go:nosplit
func _uints(p *[]byte) {
	ptr := unsafe.Pointer((*sliceHeader)(unsafe.Pointer(p)).Data)
	*((*Field)(ptr)) = Uints(*(*string)(unsafe.Add(ptr, _sizeOfField)),
		*(*[]uint)(*(*unsafe.Pointer)(unsafe.Add(ptr, _sizeOfField+_sizeOfString+_sizeOfPtr))))
}

//go:noinline
//go:nosplit
func _uint64(p *[]byte) {
	ptr := unsafe.Pointer((*sliceHeader)(unsafe.Pointer(p)).Data)
	*((*Field)(ptr)) = Uint64(*(*string)(unsafe.Add(ptr, _sizeOfField)),
		*(*uint64)(*(*unsafe.Pointer)(unsafe.Add(ptr, _sizeOfField+_sizeOfString+_sizeOfPtr))))
}

//go:noinline
//go:nosplit
func _uint64p(p *[]byte) {
	ptr := unsafe.Pointer((*sliceHeader)(unsafe.Pointer(p)).Data)
	*((*Field)(ptr)) = Uint64p(*(*string)(unsafe.Add(ptr, _sizeOfField)),
		(*uint64)(*(*unsafe.Pointer)(unsafe.Add(ptr, _sizeOfField+_sizeOfString+_sizeOfPtr))))
}

//go:noinline
//go:nosplit
func _uint64s(p *[]byte) {
	ptr := unsafe.Pointer((*sliceHeader)(unsafe.Pointer(p)).Data)
	*((*Field)(ptr)) = Uint64s(*(*string)(unsafe.Add(ptr, _sizeOfField)),
		*(*[]uint64)(*(*unsafe.Pointer)(unsafe.Add(ptr, _sizeOfField+_sizeOfString+_sizeOfPtr))))
}

//go:noinline
//go:nosplit
func _uint32(p *[]byte) {
	ptr := unsafe.Pointer((*sliceHeader)(unsafe.Pointer(p)).Data)
	*((*Field)(ptr)) = Uint32(*(*string)(unsafe.Add(ptr, _sizeOfField)),
		*(*uint32)(*(*unsafe.Pointer)(unsafe.Add(ptr, _sizeOfField+_sizeOfString+_sizeOfPtr))))
}

//go:noinline
//go:nosplit
func _uint32p(p *[]byte) {
	ptr := unsafe.Pointer((*sliceHeader)(unsafe.Pointer(p)).Data)
	*((*Field)(ptr)) = Uint32p(*(*string)(unsafe.Add(ptr, _sizeOfField)),
		(*uint32)(*(*unsafe.Pointer)(unsafe.Add(ptr, _sizeOfField+_sizeOfString+_sizeOfPtr))))
}

//go:noinline
//go:nosplit
func _uint32s(p *[]byte) {
	ptr := unsafe.Pointer((*sliceHeader)(unsafe.Pointer(p)).Data)
	*((*Field)(ptr)) = Uint32s(*(*string)(unsafe.Add(ptr, _sizeOfField)),
		*(*[]uint32)(*(*unsafe.Pointer)(unsafe.Add(ptr, _sizeOfField+_sizeOfString+_sizeOfPtr))))
}

//go:noinline
//go:nosplit
func _uint16(p *[]byte) {
	ptr := unsafe.Pointer((*sliceHeader)(unsafe.Pointer(p)).Data)
	*((*Field)(ptr)) = Uint16(*(*string)(unsafe.Add(ptr, _sizeOfField)),
		*(*uint16)(*(*unsafe.Pointer)(unsafe.Add(ptr, _sizeOfField+_sizeOfString+_sizeOfPtr))))
}

//go:noinline
//go:nosplit
func _uint16p(p *[]byte) {
	ptr := unsafe.Pointer((*sliceHeader)(unsafe.Pointer(p)).Data)
	*((*Field)(ptr)) = Uint16p(*(*string)(unsafe.Add(ptr, _sizeOfField)),
		(*uint16)(*(*unsafe.Pointer)(unsafe.Add(ptr, _sizeOfField+_sizeOfString+_sizeOfPtr))))
}

//go:noinline
//go:nosplit
func _uint16s(p *[]byte) {
	ptr := unsafe.Pointer((*sliceHeader)(unsafe.Pointer(p)).Data)
	*((*Field)(ptr)) = Uint16s(*(*string)(unsafe.Add(ptr, _sizeOfField)),
		*(*[]uint16)(*(*unsafe.Pointer)(unsafe.Add(ptr, _sizeOfField+_sizeOfString+_sizeOfPtr))))
}

//go:noinline
//go:nosplit
func _uint8(p *[]byte) {
	ptr := unsafe.Pointer((*sliceHeader)(unsafe.Pointer(p)).Data)
	*((*Field)(ptr)) = Uint8(*(*string)(unsafe.Add(ptr, _sizeOfField)),
		*(*uint8)(*(*unsafe.Pointer)(unsafe.Add(ptr, _sizeOfField+_sizeOfString+_sizeOfPtr))))
}

//go:noinline
//go:nosplit
func _uint8p(p *[]byte) {
	ptr := unsafe.Pointer((*sliceHeader)(unsafe.Pointer(p)).Data)
	*((*Field)(ptr)) = Uint8p(*(*string)(unsafe.Add(ptr, _sizeOfField)),
		(*uint8)(*(*unsafe.Pointer)(unsafe.Add(ptr, _sizeOfField+_sizeOfString+_sizeOfPtr))))
}

//go:noinline
//go:nosplit
func _binary(p *[]byte) {
	ptr := unsafe.Pointer((*sliceHeader)(unsafe.Pointer(p)).Data)
	*((*Field)(ptr)) = Binary(*(*string)(unsafe.Add(ptr, _sizeOfField)),
		*(*[]byte)(*(*unsafe.Pointer)(unsafe.Add(ptr, _sizeOfField+_sizeOfString+_sizeOfPtr))))
}

//go:noinline
//go:nosplit
func _uintptr(p *[]byte) {
	ptr := unsafe.Pointer((*sliceHeader)(unsafe.Pointer(p)).Data)
	*((*Field)(ptr)) = Uintptr(*(*string)(unsafe.Add(ptr, _sizeOfField)),
		*(*uintptr)(*(*unsafe.Pointer)(unsafe.Add(ptr, _sizeOfField+_sizeOfString+_sizeOfPtr))))
}

//go:noinline
//go:nosplit
func _uintptrp(p *[]byte) {
	ptr := unsafe.Pointer((*sliceHeader)(unsafe.Pointer(p)).Data)
	*((*Field)(ptr)) = Uintptrp(*(*string)(unsafe.Add(ptr, _sizeOfField)),
		(*uintptr)(*(*unsafe.Pointer)(unsafe.Add(ptr, _sizeOfField+_sizeOfString+_sizeOfPtr))))
}

//go:noinline
//go:nosplit
func _uintptrs(p *[]byte) {
	ptr := unsafe.Pointer((*sliceHeader)(unsafe.Pointer(p)).Data)
	*((*Field)(ptr)) = Uintptrs(*(*string)(unsafe.Add(ptr, _sizeOfField)),
		*(*[]uintptr)(*(*unsafe.Pointer)(unsafe.Add(ptr, _sizeOfField+_sizeOfString+_sizeOfPtr))))
}

//go:noinline
//go:nosplit
func _time(p *[]byte) {
	ptr := unsafe.Pointer((*sliceHeader)(unsafe.Pointer(p)).Data)
	*((*Field)(ptr)) = Time(*(*string)(unsafe.Add(ptr, _sizeOfField)),
		*(*time.Time)(*(*unsafe.Pointer)(unsafe.Add(ptr, _sizeOfField+_sizeOfString+_sizeOfPtr))))
}

//go:noinline
//go:nosplit
func _timep(p *[]byte) {
	ptr := unsafe.Pointer((*sliceHeader)(unsafe.Pointer(p)).Data)
	*((*Field)(ptr)) = Timep(*(*string)(unsafe.Add(ptr, _sizeOfField)),
		(*time.Time)(*(*unsafe.Pointer)(unsafe.Add(ptr, _sizeOfField+_sizeOfString+_sizeOfPtr))))
}

//go:noinline
//go:nosplit
func _times(p *[]byte) {
	ptr := unsafe.Pointer((*sliceHeader)(unsafe.Pointer(p)).Data)
	*((*Field)(ptr)) = Times(*(*string)(unsafe.Add(ptr, _sizeOfField)),
		*(*[]time.Time)(*(*unsafe.Pointer)(unsafe.Add(ptr, _sizeOfField+_sizeOfString+_sizeOfPtr))))
}

//go:noinline
//go:nosplit
func _duration(p *[]byte) {
	ptr := unsafe.Pointer((*sliceHeader)(unsafe.Pointer(p)).Data)
	*((*Field)(ptr)) = Duration(*(*string)(unsafe.Add(ptr, _sizeOfField)),
		*(*time.Duration)(*(*unsafe.Pointer)(unsafe.Add(ptr, _sizeOfField+_sizeOfString+_sizeOfPtr))))
}

//go:noinline
//go:nosplit
func _durationp(p *[]byte) {
	ptr := unsafe.Pointer((*sliceHeader)(unsafe.Pointer(p)).Data)
	*((*Field)(ptr)) = Durationp(*(*string)(unsafe.Add(ptr, _sizeOfField)),
		(*time.Duration)(*(*unsafe.Pointer)(unsafe.Add(ptr, _sizeOfField+_sizeOfString+_sizeOfPtr))))
}

//go:noinline
//go:nosplit
func _durations(p *[]byte) {
	ptr := unsafe.Pointer((*sliceHeader)(unsafe.Pointer(p)).Data)
	*((*Field)(ptr)) = Durations(*(*string)(unsafe.Add(ptr, _sizeOfField)),
		*(*[]time.Duration)(*(*unsafe.Pointer)(unsafe.Add(ptr, _sizeOfField+_sizeOfString+_sizeOfPtr))))
}

//go:noinline
//go:nosplit
func _namedError(p *[]byte) {
	ptr := unsafe.Pointer((*sliceHeader)(unsafe.Pointer(p)).Data)
	// error is an interface so there is no need for temporal usafe.Pointer
	*((*Field)(ptr)) = NamedError(*(*string)(unsafe.Add(ptr, _sizeOfField)),
		*(*error)(unsafe.Add(ptr, _sizeOfField+_sizeOfString)))
}

//go:noinline
//go:nosplit
func _errors(p *[]byte) {
	ptr := unsafe.Pointer((*sliceHeader)(unsafe.Pointer(p)).Data)
	*((*Field)(ptr)) = Errors(*(*string)(unsafe.Add(ptr, _sizeOfField)),
		*(*[]error)(*(*unsafe.Pointer)(unsafe.Add(ptr, _sizeOfField+_sizeOfString+_sizeOfPtr))))
}

//go:noinline
func _stringer(p *[]byte) {
	ptr := unsafe.Pointer((*sliceHeader)(unsafe.Pointer(p)).Data)
	// fmt.Stringer is an interface so there is no need for temporal usafe.Pointer
	*((*Field)(ptr)) = Stringer(*(*string)(unsafe.Add(ptr, _sizeOfField)),
		*(*fmt.Stringer)(unsafe.Add(ptr, _sizeOfField+_sizeOfString)))
}

//go:noinline
func _reflect(p *[]byte) {
	ptr := unsafe.Pointer((*sliceHeader)(unsafe.Pointer(p)).Data)
	// any is an interface so there is no need for temporal usafe.Pointer
	*((*Field)(ptr)) = Reflect(*(*string)(unsafe.Add(ptr, _sizeOfField)),
		*(*any)(unsafe.Add(ptr, _sizeOfField+_sizeOfString)))
}

type sliceHeader = reflect.SliceHeader

var _sizeOfField = unsafe.Sizeof(Field{})

const _sizeOfString = 16
const _sizeOfPtr = 8

// 64 is _sizeOfField
// 2*_sizeOfPtr is the interface
const _minimumSize = 2*_sizeOfPtr + _sizeOfString + 64
