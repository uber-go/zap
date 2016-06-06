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
)

type fieldType int

const (
	unknownType fieldType = iota
	boolType
	floatType
	intType
	int64Type
	stringType
	marshalerType
	objectType
	stringerType
)

// A Field is a deferred marshaling operation used to add a key-value pair to
// a logger's context. Keys and values are appropriately escaped for the current
// encoding scheme (e.g., JSON).
type Field struct {
	key       string
	fieldType fieldType
	ival      int64
	str       string
	obj       interface{}
}

// Bool constructs a Field with the given key and value.
func Bool(key string, val bool) Field {
	var ival int64
	if val {
		ival = 1
	}

	return Field{key: key, fieldType: boolType, ival: ival}
}

// Float64 constructs a Field with the given key and value. The floating-point
// value is encoded using strconv.FormatFloat's 'g' option (exponential notation
// for large exponents, grade-school notation otherwise).
func Float64(key string, val float64) Field {
	return Field{key: key, fieldType: floatType, ival: int64(math.Float64bits(val))}
}

// Int constructs a Field with the given key and value.
func Int(key string, val int) Field {
	return Field{key: key, fieldType: intType, ival: int64(val)}
}

// Int64 constructs a Field with the given key and value.
func Int64(key string, val int64) Field {
	return Field{key: key, fieldType: int64Type, ival: val}
}

// String constructs a Field with the given key and value.
func String(key string, val string) Field {
	return Field{key: key, fieldType: stringType, str: val}
}

// Stringer constructs a Field with the given key and value. The value
// is the result of the String method.
func Stringer(key string, val fmt.Stringer) Field {
	return Field{key: key, fieldType: stringerType, obj: val}
}

// Time constructs a Field with the given key and value. It represents a
// time.Time as nanoseconds since epoch.
func Time(key string, val time.Time) Field {
	return Int64(key, val.UnixNano())
}

// Error constructs a Field that stores err.Error() under the key "error". This is
// just a convenient shortcut for a common pattern - apart from saving a few
// keystrokes, it's no different from using zap.String.
func Error(err error) Field {
	return String("error", err.Error())
}

// Stack constructs a Field that stores a stacktrace of the current goroutine
// under the key "stacktrace". Keep in mind that taking a stacktrace is
// extremely expensive (relatively speaking); this function both makes an
// allocation and takes ~10 microseconds.
func Stack() Field {
	// Try to avoid allocating a buffer.
	enc := newJSONEncoder()
	bs := enc.bytes[:cap(enc.bytes)]
	// Returning the stacktrace as a string costs an allocation, but saves us
	// from expanding the Field union struct to include a byte slice. Since
	// taking a stacktrace is already so expensive (~10us), the extra allocation
	// is okay.
	field := String("stacktrace", takeStacktrace(bs, false))
	enc.Free()
	return field
}

// Duration constructs a Field with the given key and value. It represents
// durations as an integer number of nanoseconds.
func Duration(key string, val time.Duration) Field {
	return Int64(key, int64(val))
}

// Marshaler constructs a field with the given key and zap.LogMarshaler. It
// provides a flexible, but still type-safe and efficient, way to add
// user-defined types to the logging context.
func Marshaler(key string, val LogMarshaler) Field {
	return Field{key: key, fieldType: marshalerType, obj: val}
}

// Object constructs a field with the given key and an arbitrary object. It uses
// an encoding-appropriate, reflection-based function to serialize nearly any
// object into the logging context, but it's relatively slow and allocation-heavy.
//
// If encoding fails (e.g., trying to serialize a map[int]string to JSON), Object
// includes the error message in the final log output.
func Object(key string, val interface{}) Field {
	return Field{key: key, fieldType: objectType, obj: val}
}

// Nest takes a key and a variadic number of Fields and creates a nested
// namespace.
func Nest(key string, fields ...Field) Field {
	return Field{key: key, fieldType: marshalerType, obj: multiFields(fields)}
}

func (f Field) addTo(kv KeyValue) error {
	switch f.fieldType {
	case boolType:
		kv.AddBool(f.key, f.ival == 1)
	case floatType:
		kv.AddFloat64(f.key, math.Float64frombits(uint64(f.ival)))
	case intType:
		kv.AddInt(f.key, int(f.ival))
	case int64Type:
		kv.AddInt64(f.key, f.ival)
	case stringType:
		kv.AddString(f.key, f.str)
	case stringerType:
		kv.AddString(f.key, f.obj.(fmt.Stringer).String())
	case marshalerType:
		return kv.AddMarshaler(f.key, f.obj.(LogMarshaler))
	case objectType:
		kv.AddObject(f.key, f.obj)
	default:
		panic(fmt.Sprintf("unknown field type found: %v", f))
	}
	return nil
}

type multiFields []Field

func (fs multiFields) MarshalLog(kv KeyValue) error {
	return addFields(kv, []Field(fs))
}

func addFields(kv KeyValue, fields []Field) error {
	var errs multiError
	for _, f := range fields {
		if err := f.addTo(kv); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return errs
	}
	return nil
}
