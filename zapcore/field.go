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
	"fmt"
	"math"
)

// A FieldType indicates which member of the Field union struct should be used
// and how it should be serialized.
type FieldType uint8

const (
	// UnknownType is the default field type. Attempting to add it to an encoder will panic.
	UnknownType FieldType = iota
	// BoolType indicates that the field carries a bool.
	BoolType
	// Float64Type indicates that the field carries a float64.
	Float64Type
	// Int64Type indicates that the field carries an int64.
	Int64Type
	// Uint64Type indicates that the field carries a uint64.
	Uint64Type
	// StringType indicates that the field carries a string.
	StringType
	// ObjectMarshalerType indicates that the field carries an ObjectMarshaler.
	ObjectMarshalerType
	// ReflectType indicates that the field carries an interface{}, which should
	// be serialized using reflection.
	ReflectType
	// StringerType indicates that the field carries a fmt.Stringer.
	StringerType
	// ErrorType indicates that the field carries an error.
	ErrorType
	// SkipType indicates that the field is a no-op.
	SkipType
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
	var err error

	switch f.Type {
	case BoolType:
		enc.AddBool(f.Key, f.Integer == 1)
	case Float64Type:
		enc.AddFloat64(f.Key, math.Float64frombits(uint64(f.Integer)))
	case Int64Type:
		enc.AddInt64(f.Key, f.Integer)
	case Uint64Type:
		enc.AddUint64(f.Key, uint64(f.Integer))
	case StringType:
		enc.AddString(f.Key, f.String)
	case StringerType:
		enc.AddString(f.Key, f.Interface.(fmt.Stringer).String())
	case ObjectMarshalerType:
		err = enc.AddObject(f.Key, f.Interface.(ObjectMarshaler))
	case ReflectType:
		err = enc.AddReflected(f.Key, f.Interface)
	case ErrorType:
		enc.AddString(f.Key, f.Interface.(error).Error())
	case SkipType:
		break
	default:
		panic(fmt.Sprintf("unknown field type: %v", f))
	}

	if err != nil {
		enc.AddString(fmt.Sprintf("%sError", f.Key), err.Error())
	}
}

// Fields wraps a slice of Fields to implement ObjectMarshaler.
type Fields []Field

// MarshalLogObject implements ObjectMarshaler.
func (fs Fields) MarshalLogObject(enc ObjectEncoder) error {
	addFields(enc, []Field(fs))
	return nil
}

func addFields(enc ObjectEncoder, fields []Field) {
	for i := range fields {
		fields[i].AddTo(enc)
	}
}
