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

type anyParams struct {
	f    *Field
	k    *string
	p    unsafe.Pointer
	anyP [16]byte
}

// Any takes a key and an arbitrary value and chooses the best way to represent
// them as a field, falling back to a reflection-based approach only if
// necessary.
//
// Since byte/uint8 and rune/int32 are aliases, Any can't differentiate between
// them. To minimize surprises, []byte values are treated as binary blobs, byte
// values are treated as uint8, and runes are always treated as integers.
func Any(key string, value interface{}) (f Field) {
	params := anyParams{
		f: &f,
		k: &key,
		// most common case
		p: (*(*[2]unsafe.Pointer)(unsafe.Pointer(&value)))[1],
		// interface case
		anyP: *(*[16]byte)(unsafe.Pointer(&value)),
	}
	switch val := value.(type) {
	case zapcore.ObjectMarshaler:
		_fixInterface(&params, unsafe.Pointer(&val))
		_object(&params)
	case zapcore.ArrayMarshaler:
		_fixInterface(&params, unsafe.Pointer(&val))
		_array(&params)
	case bool:
		_bool(&params)
	case *bool:
		_boolp(&params)
	case []bool:
		_bools(&params)
	case complex128:
		_complex128(&params)
	case *complex128:
		_complex128p(&params)
	case []complex128:
		_complex128s(&params)
	case complex64:
		_complex64(&params)
	case *complex64:
		_complex64p(&params)
	case []complex64:
		_complex64s(&params)
	case float64:
		_float64(&params)
	case *float64:
		_float64p(&params)
	case []float64:
		_float64s(&params)
	case float32:
		_float32(&params)
	case *float32:
		_float32p(&params)
	case []float32:
		_float32s(&params)
	case int:
		_int(&params)
	case *int:
		_intp(&params)
	case []int:
		_ints(&params)
	case int64:
		_int64(&params)
	case *int64:
		_int64p(&params)
	case []int64:
		_int64s(&params)
	case int32:
		_int32(&params)
	case *int32:
		_int32p(&params)
	case []int32:
		_int32s(&params)
	case int16:
		_int16(&params)
	case *int16:
		_int16p(&params)
	case []int16:
		_int16s(&params)
	case int8:
		_int8(&params)
	case *int8:
		_int8p(&params)
	case []int8:
		_int8s(&params)
	case string:
		_string(&params)
	case *string:
		_stringp(&params)
	case []string:
		_strings(&params)
	case uint:
		_uint(&params)
	case *uint:
		_uintp(&params)
	case []uint:
		_uints(&params)
	case uint64:
		_uint64(&params)
	case *uint64:
		_uint64p(&params)
	case []uint64:
		_uint64s(&params)
	case uint32:
		_uint32(&params)
	case *uint32:
		_uint32p(&params)
	case []uint32:
		_uint32s(&params)
	case uint16:
		_uint16(&params)
	case *uint16:
		_uint16p(&params)
	case []uint16:
		_uint16s(&params)
	case uint8:
		_uint8(&params)
	case *uint8:
		_uint8p(&params)
	case []byte:
		_binary(&params)
	case uintptr:
		_uintptr(&params)
	case *uintptr:
		_uintptrp(&params)
	case []uintptr:
		_uintptrs(&params)
	case time.Time:
		_time(&params)
	case *time.Time:
		_timep(&params)
	case []time.Time:
		_times(&params)
	case time.Duration:
		_duration(&params)
	case *time.Duration:
		_durationp(&params)
	case []time.Duration:
		_durations(&params)
	case error:
		_fixInterface(&params, unsafe.Pointer(&val))
		_namedError(&params)
	case []error:
		_errors(&params)
	case fmt.Stringer:
		_fixInterface(&params, unsafe.Pointer(&val))
		_stringer(&params)
	default:
		_reflect(&params)
	}

	return
}

// _fixInterface makes sure that first 8 bytes are the correct object type. This happens because when we do any(10), the
// interface would look like:
// [ <pointer_to_int_type>, <pointer_to_object> ]
// but what if the object is an interface, so when we do a := any(errors.New("some error")) and then we do type
// assertion a.(error), this actually changes the first 8 bytes.
func _fixInterface(p *anyParams, valPtr unsafe.Pointer) {
	copy(p.anyP[:], (*(*[8]byte)(valPtr))[:])
}

//go:noinline
func _object(p *anyParams) {
	*p.f = Object(*p.k, *(*zapcore.ObjectMarshaler)(unsafe.Pointer(&p.anyP)))
}

//go:noinline
func _array(p *anyParams) {
	*p.f = Array(*p.k, *(*zapcore.ArrayMarshaler)(unsafe.Pointer(&p.anyP)))
}

//go:noinline
func _bool(p *anyParams) {
	*p.f = Bool(*p.k, *(*bool)(p.p))
}

//go:noinline
func _boolp(p *anyParams) {
	*p.f = Boolp(*p.k, (*bool)(p.p))
}

//go:noinline
func _bools(p *anyParams) {
	*p.f = Bools(*p.k, *(*[]bool)(p.p))
}

//go:noinline
func _complex128(p *anyParams) {
	*p.f = Complex128(*p.k, *(*complex128)(p.p))
}

//go:noinline
func _complex128p(p *anyParams) {
	*p.f = Complex128p(*p.k, (*complex128)(p.p))
}

//go:noinline
func _complex128s(p *anyParams) {
	*p.f = Complex128s(*p.k, *(*[]complex128)(p.p))
}

//go:noinline
func _complex64(p *anyParams) {
	*p.f = Complex64(*p.k, *(*complex64)(p.p))
}

//go:noinline
func _complex64p(p *anyParams) {
	*p.f = Complex64p(*p.k, (*complex64)(p.p))
}

//go:noinline
func _complex64s(p *anyParams) {
	*p.f = Complex64s(*p.k, *(*[]complex64)(p.p))
}

//go:noinline
func _float64(p *anyParams) {
	*p.f = Float64(*p.k, *(*float64)(p.p))
}

//go:noinline
func _float64p(p *anyParams) {
	*p.f = Float64p(*p.k, (*float64)(p.p))
}

//go:noinline
func _float64s(p *anyParams) {
	*p.f = Float64s(*p.k, *(*[]float64)(p.p))
}

//go:noinline
func _float32(p *anyParams) {
	*p.f = Float32(*p.k, *(*float32)(p.p))
}

//go:noinline
func _float32p(p *anyParams) {
	*p.f = Float32p(*p.k, (*float32)(p.p))
}

//go:noinline
func _float32s(p *anyParams) {
	*p.f = Float32s(*p.k, *(*[]float32)(p.p))
}

//go:noinline
func _int(p *anyParams) {
	*p.f = Int(*p.k, *(*int)(p.p))
}

//go:noinline
func _intp(p *anyParams) {
	*p.f = Intp(*p.k, (*int)(p.p))
}

//go:noinline
func _ints(p *anyParams) {
	*p.f = Ints(*p.k, *(*[]int)(p.p))
}

//go:noinline
func _int64(p *anyParams) {
	*p.f = Int64(*p.k, *(*int64)(p.p))
}

//go:noinline
func _int64p(p *anyParams) {
	*p.f = Int64p(*p.k, (*int64)(p.p))
}

//go:noinline
func _int64s(p *anyParams) {
	*p.f = Int64s(*p.k, *(*[]int64)(p.p))
}

//go:noinline
func _int32(p *anyParams) {
	*p.f = Int32(*p.k, *(*int32)(p.p))
}

//go:noinline
func _int32p(p *anyParams) {
	*p.f = Int32p(*p.k, (*int32)(p.p))
}

//go:noinline
func _int32s(p *anyParams) {
	*p.f = Int32s(*p.k, *(*[]int32)(p.p))
}

//go:noinline
func _int16(p *anyParams) {
	*p.f = Int16(*p.k, *(*int16)(p.p))
}

//go:noinline
func _int16p(p *anyParams) {
	*p.f = Int16p(*p.k, (*int16)(p.p))
}

//go:noinline
func _int16s(p *anyParams) {
	*p.f = Int16s(*p.k, *(*[]int16)(p.p))
}

//go:noinline
func _int8(p *anyParams) {
	*p.f = Int8(*p.k, *(*int8)(p.p))
}

//go:noinline

func _int8p(p *anyParams) {
	*p.f = Int8p(*p.k, (*int8)(p.p))
}

//go:noinline
func _int8s(p *anyParams) {
	*p.f = Int8s(*p.k, *(*[]int8)(p.p))
}

//go:noinline
func _string(p *anyParams) {
	*p.f = String(*p.k, *(*string)(p.p))
}

//go:noinline

func _stringp(p *anyParams) {
	*p.f = Stringp(*p.k, (*string)(p.p))
}

//go:noinline
func _strings(p *anyParams) {
	*p.f = Strings(*p.k, *(*[]string)(p.p))
}

//go:noinline
func _uint(p *anyParams) {
	*p.f = Uint(*p.k, *(*uint)(p.p))
}

//go:noinline
func _uintp(p *anyParams) {
	*p.f = Uintp(*p.k, (*uint)(p.p))
}

//go:noinline
func _uints(p *anyParams) {
	*p.f = Uints(*p.k, *(*[]uint)(p.p))
}

//go:noinline
func _uint64(p *anyParams) {
	*p.f = Uint64(*p.k, *(*uint64)(p.p))
}

//go:noinline
func _uint64p(p *anyParams) {
	*p.f = Uint64p(*p.k, (*uint64)(p.p))
}

//go:noinline
func _uint64s(p *anyParams) {
	*p.f = Uint64s(*p.k, *(*[]uint64)(p.p))
}

//go:noinline
func _uint32(p *anyParams) {
	*p.f = Uint32(*p.k, *(*uint32)(p.p))
}

//go:noinline
func _uint32p(p *anyParams) {
	*p.f = Uint32p(*p.k, (*uint32)(p.p))
}

//go:noinline
func _uint32s(p *anyParams) {
	*p.f = Uint32s(*p.k, *(*[]uint32)(p.p))
}

//go:noinline
func _uint16(p *anyParams) {
	*p.f = Uint16(*p.k, *(*uint16)(p.p))
}

//go:noinline
func _uint16p(p *anyParams) {
	*p.f = Uint16p(*p.k, (*uint16)(p.p))
}

//go:noinline
func _uint16s(p *anyParams) {
	*p.f = Uint16s(*p.k, *(*[]uint16)(p.p))
}

//go:noinline
func _uint8(p *anyParams) {
	*p.f = Uint8(*p.k, *(*uint8)(p.p))
}

//go:noinline
func _uint8p(p *anyParams) {
	*p.f = Uint8p(*p.k, (*uint8)(p.p))
}

//go:noinline
func _binary(p *anyParams) {
	*p.f = Binary(*p.k, *(*[]byte)(p.p))
}

//go:noinline
func _uintptr(p *anyParams) {
	*p.f = Uintptr(*p.k, *(*uintptr)(p.p))
}

//go:noinline
func _uintptrp(p *anyParams) {
	*p.f = Uintptrp(*p.k, (*uintptr)(p.p))
}

//go:noinline
func _uintptrs(p *anyParams) {
	*p.f = Uintptrs(*p.k, *(*[]uintptr)(p.p))
}

//go:noinline
func _time(p *anyParams) {
	*p.f = Time(*p.k, *(*time.Time)(p.p))
}

//go:noinline
func _timep(p *anyParams) {
	*p.f = Timep(*p.k, (*time.Time)(p.p))
}

//go:noinline
func _times(p *anyParams) {
	*p.f = Times(*p.k, *(*[]time.Time)(p.p))
}

//go:noinline
func _duration(p *anyParams) {
	*p.f = Duration(*p.k, *(*time.Duration)(p.p))
}

//go:noinline
func _durationp(p *anyParams) {
	*p.f = Durationp(*p.k, (*time.Duration)(p.p))
}

//go:noinline
func _durations(p *anyParams) {
	*p.f = Durations(*p.k, *(*[]time.Duration)(p.p))
}

//go:noinline
func _namedError(p *anyParams) {
	*p.f = NamedError(*p.k, *(*error)(unsafe.Pointer(&p.anyP)))
}

//go:noinline
func _errors(p *anyParams) {
	*p.f = Errors(*p.k, *(*[]error)(p.p))
}

//go:noinline
func _stringer(p *anyParams) {
	*p.f = Stringer(*p.k, *(*fmt.Stringer)(unsafe.Pointer(&p.anyP)))
}

//go:noinline
func _reflect(p *anyParams) {
	*p.f = Reflect(*p.k, *(*any)(unsafe.Pointer(&p.anyP)))
}
