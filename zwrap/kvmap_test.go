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

import (
	"errors"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/uber-go/zap"
)

type unloggable struct{}

func (unloggable) MarshalLog(kv zap.KeyValue) error {
	return errors.New("not loggable")
}

type loggable struct{}

func (l loggable) MarshalLog(kv zap.KeyValue) error {
	kv.AddString("loggable", "yes")
	return nil
}

func TestKeyValueMapAdd(t *testing.T) {
	arbitraryObj := map[string]interface{}{
		"foo": "bar",
		"baz": 5,
	}

	kv := KeyValueMap{}
	kv.AddBool("b", true)
	kv.AddFloat64("f64", 1.56)
	kv.AddInt("int", 5)
	kv.AddInt64("i64", math.MaxInt64)
	kv.AddUintptr("uintptr", uintptr(0xdeadbeef))
	kv.AddString("s", "string")

	assert.NoError(t, kv.AddObject("obj", arbitraryObj), "AddObject failed")
	assert.NoError(t, kv.AddMarshaler("m1", loggable{}), "AddMarshaler failed")
	assert.NoError(t, kv.Nest("m2", loggable{}.MarshalLog), "Nest failed")

	want := KeyValueMap{
		"b":       true,
		"f64":     1.56,
		"int":     5,
		"i64":     int64(math.MaxInt64),
		"uintptr": uintptr(0xdeadbeef),
		"s":       "string",
		"obj":     arbitraryObj,
		"m1": KeyValueMap{
			"loggable": "yes",
		},
		"m2": KeyValueMap{
			"loggable": "yes",
		},
	}
	assert.Equal(t, want, kv, "Unexpected result")
}

func TestKeyValueMapAddFails(t *testing.T) {
	kv := KeyValueMap{}

	assert.Error(t, kv.AddMarshaler("m1", unloggable{}), "AddMarshaler should fail")
	assert.Error(t, kv.Nest("m2", unloggable{}.MarshalLog), "Nest should fail")
	assert.Equal(t, KeyValueMap{
		"m1": KeyValueMap{},
		"m2": KeyValueMap{},
	}, kv, "Empty values on errors")
}
