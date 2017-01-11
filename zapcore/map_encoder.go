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

// AddBool adds the value under the specified key to the map.
func (m MapObjectEncoder) AddBool(k string, v bool) { m[k] = v }

// AddFloat64 adds the value under the specified key to the map.
func (m MapObjectEncoder) AddFloat64(k string, v float64) { m[k] = v }

// AddInt64 adds the value under the specified key to the map.
func (m MapObjectEncoder) AddInt64(k string, v int64) { m[k] = v }

// AddUint64 adds the value under the specified key to the map.
func (m MapObjectEncoder) AddUint64(k string, v uint64) { m[k] = v }

// AddReflected adds the value under the specified key to the map.
func (m MapObjectEncoder) AddReflected(k string, v interface{}) error {
	m[k] = v
	return nil
}

// AddString adds the value under the specified key to the map.
func (m MapObjectEncoder) AddString(k string, v string) { m[k] = v }

// AddObject adds the value under the specified key to the map.
func (m MapObjectEncoder) AddObject(k string, v ObjectMarshaler) error {
	newMap := make(MapObjectEncoder)
	m[k] = newMap
	return v.MarshalLogObject(newMap)
}
