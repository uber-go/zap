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
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMapEncoderAdd(t *testing.T) {
	arbitraryObj := map[string]interface{}{
		"foo": "bar",
		"baz": 5,
	}

	enc := make(MapObjectEncoder)
	enc.AddBool("b", true)
	enc.AddFloat64("f64", 1.56)
	enc.AddInt("int", 5)
	enc.AddInt64("i64", math.MaxInt64)
	enc.AddUintptr("uintptr", uintptr(0xdeadbeef))
	enc.AddString("s", "string")

	assert.NoError(t, enc.AddReflected("reflect", arbitraryObj), "Expected AddReflected to succeed.")
	assert.NoError(t, enc.AddObject("object", loggable{true}), "Expected AddObject to succeed.")

	want := MapObjectEncoder{
		"b":       true,
		"f64":     1.56,
		"int":     5,
		"i64":     int64(math.MaxInt64),
		"uintptr": uintptr(0xdeadbeef),
		"s":       "string",
		"reflect": arbitraryObj,
		"object": MapObjectEncoder{
			"loggable": "yes",
		},
	}
	assert.Equal(t, want, enc, "Encoder's final state is unexpected.")
}

func TestKeyValueMapAddFails(t *testing.T) {
	enc := make(MapObjectEncoder)
	assert.Error(t, enc.AddObject("object", loggable{false}), "Expected AddObject to fail.")
	assert.Equal(t, MapObjectEncoder{"object": MapObjectEncoder{}}, enc, "Expected encoder to use empty values on errors.")
}
