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

import "time"

// A Field is a deferred marshaling operation used to add a key-value pair to
// a logger's context. Keys and values are appropriately escaped for the current
// encoding scheme (e.g., JSON).
type Field interface {
	addTo(encoder) error
}

// A FieldCloser closes a nested field.
type FieldCloser interface {
	CloseField()
}

// Bool constructs a Field with the given key and value.
func Bool(key string, val bool) Field {
	return boolField{key, val}
}

// Float64 constructs a Field with the given key and value. The floating-point
// value is encoded using strconv.FormatFloat's 'g' option (exponential notation
// for large exponents, grade-school notation otherwise).
func Float64(key string, val float64) Field {
	return float64Field{key, val}
}

// Int constructs a Field with the given key and value.
func Int(key string, val int) Field {
	return int64Field{key, int64(val)}
}

// Int64 constructs a Field with the given key and value.
func Int64(key string, val int64) Field {
	return int64Field{key, val}
}

// String constructs a Field with the given key and value.
func String(key string, val string) Field {
	return stringField{key, val}
}

// Time constructs a Field with the given key and value. It represents a
// time.Time as nanoseconds since epoch.
func Time(key string, val time.Time) Field {
	return timeField{key, val}
}

// Err constructs a Field that stores err.Error() under the key "error".
func Err(err error) Field {
	return stringField{"error", err.Error()}
}

// Duration constructs a Field with the given key and value. It represents
// durations as an integer number of nanoseconds.
func Duration(key string, val time.Duration) Field {
	return int64Field{key, int64(val)}
}

// Object constructs a field with the given key and zap.Marshaler. It provides a
// flexible, but still type-safe and efficient, way to add user-defined types to
// the logging context.
func Object(key string, val Marshaler) Field {
	return marshalerField{key, val}
}

// Nest takes a key and a variadic number of Fields and creates a nested
// namespace.
func Nest(key string, fields ...Field) Field {
	return nestedField{key, fields}
}

type boolField struct {
	key string
	val bool
}

func (b boolField) addTo(enc encoder) error {
	enc.AddBool(b.key, b.val)
	return nil
}

type float64Field struct {
	key string
	val float64
}

func (f float64Field) addTo(enc encoder) error {
	enc.AddFloat64(f.key, f.val)
	return nil
}

type int64Field struct {
	key string
	val int64
}

func (i int64Field) addTo(enc encoder) error {
	enc.AddInt64(i.key, i.val)
	return nil
}

type stringField struct {
	key string
	val string
}

func (s stringField) addTo(enc encoder) error {
	enc.AddString(s.key, s.val)
	return nil
}

type timeField struct {
	key string
	val time.Time
}

func (t timeField) addTo(enc encoder) error {
	enc.AddTime(t.key, t.val)
	return nil
}

type marshalerField struct {
	key string
	val Marshaler
}

func (m marshalerField) addTo(enc encoder) error {
	closer := enc.Nest(m.key)
	err := m.val.MarshalLog(enc)
	closer.CloseField()
	return err
}

type nestedField struct {
	key  string
	vals []Field
}

func (n nestedField) addTo(enc encoder) error {
	closer := enc.Nest(n.key)
	var errs multiError
	for _, f := range n.vals {
		if err := f.addTo(enc); err != nil {
			errs = append(errs, err)
		}
	}
	closer.CloseField()
	if len(errs) > 0 {
		return errs
	}
	return nil
}
