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

func TestMapObjectEncoderAdd(t *testing.T) {
	arbitraryObj := map[string]interface{}{
		"foo": "bar",
		"baz": 5,
	}

	enc := make(MapObjectEncoder)
	// ObjectEncoder methods.
	enc.AddBool("b", true)
	enc.AddFloat64("f64", 1.56)
	enc.AddInt64("i64", math.MaxInt64)
	enc.AddUint64("uint64", 42)
	enc.AddString("s", "string")
	assert.NoError(t, enc.AddReflected("reflect", arbitraryObj), "Expected AddReflected to succeed.")
	assert.NoError(t, enc.AddObject("object", loggable{true}), "Expected AddObject to succeed.")
	// Array types.
	assert.NoError(t, enc.AddArray("array", ArrayMarshalerFunc(func(arr ArrayEncoder) error {
		arr.AppendBool(true)
		arr.AppendBool(false)
		arr.AppendBool(true)
		return nil
	})), "Expected AddArray to succeed.")
	assert.NoError(t, enc.AddArray("arrays-of-arrays", ArrayMarshalerFunc(func(enc ArrayEncoder) error {
		enc.AppendArray(ArrayMarshalerFunc(func(e ArrayEncoder) error {
			e.AppendBool(true)
			return nil
		}))
		return nil
	})), "Expected AddArray to succeed.")
	// Nested objects and arrays.
	assert.NoError(t, enc.AddArray("turduckens", turduckens(2)), "Expected AddObject to succeed.")
	assert.NoError(t, enc.AddObject("turducken", turducken{}), "Expected AddObject to succeed.")

	wantTurducken := MapObjectEncoder{
		"ducks": []interface{}{
			MapObjectEncoder{"in": "chicken"},
			MapObjectEncoder{"in": "chicken"},
		},
	}
	want := MapObjectEncoder{
		"b":       true,
		"f64":     1.56,
		"i64":     int64(math.MaxInt64),
		"uint64":  uint64(42),
		"s":       "string",
		"reflect": arbitraryObj,
		"object": MapObjectEncoder{
			"loggable": "yes",
		},
		"array":            []interface{}{true, false, true},
		"arrays-of-arrays": []interface{}{[]interface{}{true}},
		"turducken":        wantTurducken,
		"turduckens":       []interface{}{wantTurducken, wantTurducken},
	}
	assert.Equal(t, want, enc, "Encoder's final state is unexpected.")
}

func TestKeyValueMapAddFails(t *testing.T) {
	enc := make(MapObjectEncoder)
	assert.Error(t, enc.AddObject("object", loggable{false}), "Expected AddObject to fail.")
	assert.Equal(t, MapObjectEncoder{"object": MapObjectEncoder{}}, enc, "Expected encoder to use empty values on errors.")
}
