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

// MapObjectEncoder is an ObjectEncoder backed by a simple
// map[string]interface{}. It's not fast enough for production use, but it's
// helpful in tests.
type MapObjectEncoder map[string]interface{}

// AddBool implements ObjectEncoder.
func (m MapObjectEncoder) AddBool(k string, v bool) { m[k] = v }

// AddFloat64 implements ObjectEncoder.
func (m MapObjectEncoder) AddFloat64(k string, v float64) { m[k] = v }

// AddInt64 implements ObjectEncoder.
func (m MapObjectEncoder) AddInt64(k string, v int64) { m[k] = v }

// AddUint64 implements ObjectEncoder.
func (m MapObjectEncoder) AddUint64(k string, v uint64) { m[k] = v }

// AddReflected implements ObjectEncoder.
func (m MapObjectEncoder) AddReflected(k string, v interface{}) error {
	m[k] = v
	return nil
}

// AddString implements ObjectEncoder.
func (m MapObjectEncoder) AddString(k string, v string) { m[k] = v }

// AddObject implements ObjectEncoder.
func (m MapObjectEncoder) AddObject(k string, v ObjectMarshaler) error {
	newMap := make(MapObjectEncoder)
	m[k] = newMap
	return v.MarshalLogObject(newMap)
}

// AddArray implements ObjectEncoder.
func (m MapObjectEncoder) AddArray(key string, v ArrayMarshaler) error {
	arr := &sliceArrayEncoder{}
	err := v.MarshalLogArray(arr)
	m[key] = arr.elems
	return err
}

// sliceArrayEncoder is an ArrayEncoder backed by a simple []interface{}. Like
// the MapObjectEncoder, it's not designed for production use.
type sliceArrayEncoder struct {
	elems []interface{}
}

func (s *sliceArrayEncoder) AppendArray(v ArrayMarshaler) error {
	enc := &sliceArrayEncoder{}
	err := v.MarshalLogArray(enc)
	s.elems = append(s.elems, enc.elems)
	return err
}

func (s *sliceArrayEncoder) AppendObject(v ObjectMarshaler) error {
	m := make(MapObjectEncoder)
	err := v.MarshalLogObject(m)
	s.elems = append(s.elems, m)
	return err
}

func (s *sliceArrayEncoder) AppendBool(v bool) {
	s.elems = append(s.elems, v)
}
