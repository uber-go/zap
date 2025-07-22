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

package zapcore

import (
	"bytes"
	"fmt"
	"math"
	"reflect"
	"time"
)

// A FieldType indicates which member of the Field union struct should be used
// and how it should be serialized.
type FieldType uint8

const (
	// UnknownType is the default field type. Attempting to add it to an encoder will panic.
	UnknownType FieldType = iota
	// ArrayMarshalerType indicates that the field carries an ArrayMarshaler.
	ArrayMarshalerType
	// ObjectMarshalerType indicates that the field carries an ObjectMarshaler.
	ObjectMarshalerType
	// BinaryType indicates that the field carries an opaque binary blob.
	BinaryType
	// BoolType indicates that the field carries a bool.
	BoolType
	// ByteStringType indicates that the field carries UTF-8 encoded bytes.
	ByteStringType
	// Complex128Type indicates that the field carries a complex128.
	Complex128Type
	// Complex64Type indicates that the field carries a complex64.
	Complex64Type
	// DurationType indicates that the field carries a time.Duration.
	DurationType
	// Float64Type indicates that the field carries a float64.
	Float64Type
	// Float32Type indicates that the field carries a float32.
	Float32Type
	// Int64Type indicates that the field carries an int64.
	Int64Type
	// Int32Type indicates that the field carries an int32.
	Int32Type
	// Int16Type indicates that the field carries an int16.
	Int16Type
	// Int8Type indicates that the field carries an int8.
	Int8Type
	// StringType indicates that the field carries a string.
	StringType
	// TimeType indicates that the field carries a time.Time that is
	// representable by a UnixNano() stored as an int64.
	TimeType
	// TimeFullType indicates that the field carries a time.Time stored as-is.
	TimeFullType
	// Uint64Type indicates that the field carries a uint64.
	Uint64Type
	// Uint32Type indicates that the field carries a uint32.
	Uint32Type
	// Uint16Type indicates that the field carries a uint16.
	Uint16Type
	// Uint8Type indicates that the field carries a uint8.
	Uint8Type
	// UintptrType indicates that the field carries a uintptr.
	UintptrType
	// ReflectType indicates that the field carries an interface{}, which should
	// be serialized using reflection.
	ReflectType
	// NamespaceType signals the beginning of an isolated namespace. All
	// subsequent fields should be added to the new namespace.
	NamespaceType
	// StringerType indicates that the field carries a fmt.Stringer.
	StringerType
	// ErrorType indicates that the field carries an error.
	ErrorType
	// SkipType indicates that the field is a no-op.
	SkipType

	// InlineMarshalerType indicates that the field carries an ObjectMarshaler
	// that should be inlined.
	InlineMarshalerType
)

// A Field is a marshaling operation used to add a key-value pair to a logger's
// context. Most fields are lazily marshaled, so it's inexpensive to add fields
// to disabled debug-level log statements.
type Field struct {
	Key       string
	Type      FieldType
	Integer   int64
	String    string
	Interface interface{}
}

// AddTo exports a field through the ObjectEncoder interface. It's primarily
// useful to library authors, and shouldn't be necessary in most applications.
func (f Field) AddTo(enc ObjectEncoder) {
	// Avoid allocating err unless needed
	var err error

	// Fast-path: handle most common types without allocations or assertions.
	switch f.Type {
	case BoolType:
		enc.AddBool(f.Key, f.Integer == 1)
		return
	case Int64Type:
		enc.AddInt64(f.Key, f.Integer)
		return
	case Int32Type:
		enc.AddInt32(f.Key, int32(f.Integer))
		return
	case Int16Type:
		enc.AddInt16(f.Key, int16(f.Integer))
		return
	case Int8Type:
		enc.AddInt8(f.Key, int8(f.Integer))
		return
	case Uint64Type:
		enc.AddUint64(f.Key, uint64(f.Integer))
		return
	case Uint32Type:
		enc.AddUint32(f.Key, uint32(f.Integer))
		return
	case Uint16Type:
		enc.AddUint16(f.Key, uint16(f.Integer))
		return
	case Uint8Type:
		enc.AddUint8(f.Key, uint8(f.Integer))
		return
	case Float64Type:
		enc.AddFloat64(f.Key, math.Float64frombits(uint64(f.Integer)))
		return
	case Float32Type:
		enc.AddFloat32(f.Key, math.Float32frombits(uint32(f.Integer)))
		return
	case StringType:
		enc.AddString(f.Key, f.String)
		return
	case DurationType:
		enc.AddDuration(f.Key, time.Duration(f.Integer))
		return
	case TimeType:
		if f.Interface != nil {
			loc, ok := f.Interface.(*time.Location)
			if ok && loc != nil {
				enc.AddTime(f.Key, time.Unix(0, f.Integer).In(loc))
			} else {
				enc.AddTime(f.Key, time.Unix(0, f.Integer))
			}
		} else {
			// Fall back to UTC if location is nil.
			enc.AddTime(f.Key, time.Unix(0, f.Integer))
		}
		return
	case NamespaceType:
		enc.OpenNamespace(f.Key)
		return
	case SkipType:
		return
	}

	// Slower path with interface usage/assertion.
	switch f.Type {
	case ArrayMarshalerType:
		if arr, ok := f.Interface.(ArrayMarshaler); ok {
			err = enc.AddArray(f.Key, arr)
		} else {
			err = fmt.Errorf("Field.Interface is not ArrayMarshaler for %s", f.Key)
		}
	case ObjectMarshalerType:
		if obj, ok := f.Interface.(ObjectMarshaler); ok {
			err = enc.AddObject(f.Key, obj)
		} else {
			err = fmt.Errorf("Field.Interface is not ObjectMarshaler for %s", f.Key)
		}
	case InlineMarshalerType:
		if inl, ok := f.Interface.(ObjectMarshaler); ok {
			err = inl.MarshalLogObject(enc)
		} else {
			err = fmt.Errorf("Field.Interface is not ObjectMarshaler for %s", f.Key)
		}
	case BinaryType:
		if b, ok := f.Interface.([]byte); ok {
			enc.AddBinary(f.Key, b)
		} else {
			err = fmt.Errorf("Field.Interface is not []byte for %s", f.Key)
		}
	case ByteStringType:
		if b, ok := f.Interface.([]byte); ok {
			enc.AddByteString(f.Key, b)
		} else {
			err = fmt.Errorf("Field.Interface is not []byte for %s", f.Key)
		}
	case Complex128Type:
		if c, ok := f.Interface.(complex128); ok {
			enc.AddComplex128(f.Key, c)
		} else {
			err = fmt.Errorf("Field.Interface is not complex128 for %s", f.Key)
		}
	case Complex64Type:
		if c, ok := f.Interface.(complex64); ok {
			enc.AddComplex64(f.Key, c)
		} else {
			err = fmt.Errorf("Field.Interface is not complex64 for %s", f.Key)
		}
	case TimeFullType:
		if t, ok := f.Interface.(time.Time); ok {
			enc.AddTime(f.Key, t)
		} else {
			err = fmt.Errorf("Field.Interface is not time.Time for %s", f.Key)
		}
	case UintptrType:
		enc.AddUintptr(f.Key, uintptr(f.Integer))
	case ReflectType:
		err = enc.AddReflected(f.Key, f.Interface)
	case StringerType:
		err = encodeStringer(f.Key, f.Interface, enc)
	case ErrorType:
		if e, ok := f.Interface.(error); ok {
			err = encodeError(f.Key, e, enc)
		} else {
			err = fmt.Errorf("Field.Interface is not error for %s", f.Key)
		}
	default:
		panic(fmt.Sprintf("unknown field type: %v", f))
	}

	if err != nil {
		// Avoid Sprintf allocation if possible
		key := f.Key
		if key == "" {
			key = "Field"
		}
		enc.AddString(key+"Error", err.Error())
	}
}

// Equals returns whether two fields are equal. For non-primitive types such as
// errors, marshalers, or reflect types, it uses reflect.DeepEqual.
func (f Field) Equals(other Field) bool {
	if f.Type != other.Type {
		return false
	}
	if f.Key != other.Key {
		return false
	}

	switch f.Type {
	case BinaryType, ByteStringType:
		// Safely handle type assertions with explicit checks
		fb, fok := f.Interface.([]byte)
		ob, ook := other.Interface.([]byte)
		if !fok || !ook {
			return false
		}
		return bytes.Equal(fb, ob)
	case ArrayMarshalerType, ObjectMarshalerType, ErrorType, ReflectType:
		// For complex types we still need DeepEqual
		// This is typically used in tests, not in hot paths
		return reflect.DeepEqual(f.Interface, other.Interface)
	default:
		// For primitive types, direct comparison is efficient
		return f == other
	}
}

func addFields(enc ObjectEncoder, fields []Field) {
	// Using direct indexing to avoid creating a copy of the Field
	for i := range fields {
		fields[i].AddTo(enc)
	}
}

func encodeStringer(key string, stringer interface{}, enc ObjectEncoder) (retErr error) {
	// Try to capture panics (from nil references or otherwise) when calling
	// the String() method, similar to https://golang.org/src/fmt/print.go#L540
	defer func() {
		if err := recover(); err != nil {
			// If it's a nil pointer, just say "<nil>". The likeliest causes are a
			// Stringer that fails to guard against nil or a nil pointer for a
			// value receiver, and in either case, "<nil>" is a nice result.
			if v := reflect.ValueOf(stringer); v.Kind() == reflect.Ptr && v.IsNil() {
				enc.AddString(key, "<nil>")
				return
			}

			retErr = fmt.Errorf("PANIC=%v", err)
		}
	}()

	enc.AddString(key, stringer.(fmt.Stringer).String())
	return nil
}