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

package zwrap

import "github.com/uber-go/zap"

// KeyValueMap implements zap.KeyValue backed by a map.
type KeyValueMap map[string]interface{}

// AddBool adds the value under the specified key to the map.
func (m KeyValueMap) AddBool(k string, v bool) { m[k] = v }

// AddFloat64 adds the value under the specified key to the map.
func (m KeyValueMap) AddFloat64(k string, v float64) { m[k] = v }

// AddInt adds the value under the specified key to the map.
func (m KeyValueMap) AddInt(k string, v int) { m[k] = v }

// AddInt64 adds the value under the specified key to the map.
func (m KeyValueMap) AddInt64(k string, v int64) { m[k] = v }

// AddObject adds the value under the specified key to the map.
func (m KeyValueMap) AddObject(k string, v interface{}) error {
	m[k] = v
	return nil
}

// AddString adds the value under the specified key to the map.
func (m KeyValueMap) AddString(k string, v string) { m[k] = v }

// AddMarshaler adds the value under the specified key to the map.
func (m KeyValueMap) AddMarshaler(k string, v zap.LogMarshaler) error {
	return m.Nest(k, v.MarshalLog)
}

// Nest builds a object and adds the value under the specified key to the map.
func (m KeyValueMap) Nest(k string, f func(zap.KeyValue) error) error {
	newMap := make(KeyValueMap)
	m[k] = newMap
	return f(newMap)
}
