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
	"sync"
	"sync/atomic"
	"time"
)

var (
	// Pool all the things.
	_boolFieldPool      = sync.Pool{New: func() interface{} { return &boolField{fieldRefCount: new(fieldRefCount)} }}
	_float64FieldPool   = sync.Pool{New: func() interface{} { return &float64Field{fieldRefCount: new(fieldRefCount)} }}
	_int64FieldPool     = sync.Pool{New: func() interface{} { return &int64Field{fieldRefCount: new(fieldRefCount)} }}
	_stringFieldPool    = sync.Pool{New: func() interface{} { return &stringField{fieldRefCount: new(fieldRefCount)} }}
	_timeFieldPool      = sync.Pool{New: func() interface{} { return &timeField{fieldRefCount: new(fieldRefCount)} }}
	_marshalerFieldPool = sync.Pool{New: func() interface{} { return &marshalerField{fieldRefCount: new(fieldRefCount)} }}
	_nestedFieldPool    = sync.Pool{New: func() interface{} { return &nestedField{fieldRefCount: new(fieldRefCount)} }}

	_msgTooManyRefs = "Field was freed with an active reference. To re-use fields, use zap.Keep (https://godoc.org/github.com/uber-common/zap/#Keep)."
	_msgFreedTwice  = "Field was freed more than once. To re-use fields, use zap.Keep (https://godoc.org/github.com/uber-common/zap/#Keep)."
)

// A FieldOption configures a field.
type FieldOption func(Field)

// Keep prevents the field from being freed after use. Fields that haven't been
// marked by Keep will be returned to a sync.Pool immediately after use (in
// NewJSON, Logger.With, or one of the leveled logging methods).
func Keep(f Field) {
	f.doNotFree()
}

// A Field is a deferred marshaling operation used to add a key-value pair to
// a logger's context. Keys and values are appropriately escaped for the current
// encoding scheme (e.g., JSON).
//
// Fields are returned to a sync.Pool immediately after they're passed to a
// logger (in NewJSON, Logger.With, Logger.Debug, or any of the other leveled
// logging methods). This means that they're not safe for re-use!
type Field interface {
	addTo(encoder) error
	doNotFree()
	incRef() int32
}

// A FieldCloser closes a nested field.
type FieldCloser interface {
	CloseField()
}

// Bool constructs a Field with the given key and value.
func Bool(key string, val bool, opts ...FieldOption) Field {
	field := _boolFieldPool.Get().(*boolField)
	field.key = key
	field.val = val
	for _, opt := range opts {
		opt(field)
	}
	field.incRef()
	return field
}

// Float64 constructs a Field with the given key and value. The floating-point
// value is encoded using strconv.FormatFloat's 'g' option (exponential notation
// for large exponents, grade-school notation otherwise).
func Float64(key string, val float64, opts ...FieldOption) Field {
	field := _float64FieldPool.Get().(*float64Field)
	field.key = key
	field.val = val
	for _, opt := range opts {
		opt(field)
	}
	field.incRef()
	return field
}

// Int constructs a Field with the given key and value.
func Int(key string, val int, opts ...FieldOption) Field {
	return Int64(key, int64(val), opts...)
}

// Int64 constructs a Field with the given key and value.
func Int64(key string, val int64, opts ...FieldOption) Field {
	field := _int64FieldPool.Get().(*int64Field)
	field.key = key
	field.val = val
	for _, opt := range opts {
		opt(field)
	}
	field.incRef()
	return field
}

// String constructs a Field with the given key and value.
func String(key string, val string, opts ...FieldOption) Field {
	field := _stringFieldPool.Get().(*stringField)
	field.key = key
	field.val = val
	for _, opt := range opts {
		opt(field)
	}
	field.incRef()
	return field
}

// Time constructs a Field with the given key and value. It represents a
// time.Time as nanoseconds since epoch.
func Time(key string, val time.Time, opts ...FieldOption) Field {
	field := _timeFieldPool.Get().(*timeField)
	field.key = key
	field.val = val
	for _, opt := range opts {
		opt(field)
	}
	field.incRef()
	return field
}

// Err constructs a Field that stores err.Error() under the key "error".
func Err(err error, opts ...FieldOption) Field {
	return String("error", err.Error(), opts...)
}

// Duration constructs a Field with the given key and value. It represents
// durations as an integer number of nanoseconds.
func Duration(key string, val time.Duration, opts ...FieldOption) Field {
	return Int64(key, int64(val), opts...)
}

// Object constructs a field with the given key and zap.Marshaler. It provides a
// flexible, but still type-safe and efficient, way to add user-defined types to
// the logging context.
func Object(key string, val Marshaler, opts ...FieldOption) Field {
	field := _marshalerFieldPool.Get().(*marshalerField)
	field.key = key
	field.val = val
	for _, opt := range opts {
		opt(field)
	}
	field.incRef()
	return field
}

// Nest takes a key and a variadic number of fields and creates a nested
// namespace.
//
// Because Nest already takes a variadic number of fields, it doesn't take any
// options. To re-use a nested field, call Keep on it manually.
func Nest(key string, fields ...Field) Field {
	field := _nestedFieldPool.Get().(*nestedField)
	field.key = key
	field.vals = fields
	field.incRef()
	return field
}

type fieldRefCount struct {
	n    int32
	keep bool
}

func (rc *fieldRefCount) incRef() int32 {
	return atomic.AddInt32(&rc.n, 1)
}

func (rc *fieldRefCount) doNotFree() {
	rc.keep = true
}

func (rc *fieldRefCount) shouldFree() bool {
	if rc.keep {
		return false
	}
	refs := atomic.AddInt32(&rc.n, -1)
	if refs > 0 {
		panic(_msgTooManyRefs)
	} else if refs < 0 {
		panic(_msgFreedTwice)
	}
	return true
}

type boolField struct {
	*fieldRefCount
	key string
	val bool
}

func (b *boolField) addTo(enc encoder) error {
	enc.AddBool(b.key, b.val)
	b.free()
	return nil
}

func (b *boolField) free() {
	if b.shouldFree() {
		_boolFieldPool.Put(b)
	}
}

type float64Field struct {
	*fieldRefCount
	key string
	val float64
}

func (f *float64Field) addTo(enc encoder) error {
	enc.AddFloat64(f.key, f.val)
	f.free()
	return nil
}

func (f *float64Field) free() {
	if f.shouldFree() {
		_float64FieldPool.Put(f)
	}
}

type int64Field struct {
	*fieldRefCount
	key string
	val int64
}

func (i *int64Field) addTo(enc encoder) error {
	enc.AddInt64(i.key, i.val)
	i.free()
	return nil
}

func (i *int64Field) free() {
	if i.shouldFree() {
		_int64FieldPool.Put(i)
	}
}

type stringField struct {
	*fieldRefCount
	key string
	val string
}

func (s *stringField) addTo(enc encoder) error {
	enc.AddString(s.key, s.val)
	s.free()
	return nil
}

func (s *stringField) free() {
	if s.shouldFree() {
		_stringFieldPool.Put(s)
	}
}

type timeField struct {
	*fieldRefCount
	key string
	val time.Time
}

func (t *timeField) addTo(enc encoder) error {
	enc.AddTime(t.key, t.val)
	t.free()
	return nil
}

func (t *timeField) free() {
	if t.shouldFree() {
		_timeFieldPool.Put(t)
	}
}

type marshalerField struct {
	*fieldRefCount
	key string
	val Marshaler
}

func (m *marshalerField) addTo(enc encoder) error {
	closer := enc.Nest(m.key)
	err := m.val.MarshalLog(enc)
	closer.CloseField()
	m.free()
	return err
}

func (m *marshalerField) free() {
	if m.shouldFree() {
		_marshalerFieldPool.Put(m)
	}
}

type nestedField struct {
	*fieldRefCount
	key  string
	vals []Field
}

func (n *nestedField) addTo(enc encoder) error {
	closer := enc.Nest(n.key)
	var errs multiError
	for _, f := range n.vals {
		if err := f.addTo(enc); err != nil {
			errs = append(errs, err)
		}
	}
	closer.CloseField()
	n.free()
	if len(errs) > 0 {
		return errs
	}
	return nil
}

func (n *nestedField) doNotFree() {
	n.fieldRefCount.doNotFree()
	for _, f := range n.vals {
		f.doNotFree()
	}
}

func (n *nestedField) free() {
	if n.shouldFree() {
		_nestedFieldPool.Put(n)
	}
}
