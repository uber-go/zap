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
	"encoding/json"
	"math/rand"
	"reflect"
	"testing"
	"testing/quick"
)

func zapEncodeString(s string) []byte {
	// Escape and quote a string using our encoder.
	var ret []byte
	withJSONEncoder(func(enc *jsonEncoder) {
		enc.safeAddString(s)
		ret = make([]byte, 0, len(enc.bytes)+2)
		ret = append(ret, '"')
		ret = append(ret, enc.bytes...)
		ret = append(ret, '"')
	})
	return ret
}

func roundTripsCorrectly(original string) bool {
	// Encode using our encoder, decode using the standard library, and assert
	// that we haven't lost any information.
	encoded := zapEncodeString(original)

	var decoded string
	err := json.Unmarshal(encoded, &decoded)
	if err != nil {
		return false
	}
	return original == decoded
}

type ASCII string

func (s ASCII) Generate(r *rand.Rand, size int) reflect.Value {
	bs := make([]byte, size)
	for i := range bs {
		bs[i] = byte(r.Intn(128))
	}
	a := ASCII(bs)
	return reflect.ValueOf(a)
}

func asciiRoundTripsCorrectly(s ASCII) bool {
	return roundTripsCorrectly(string(s))
}

func TestJSONQuick(t *testing.T) {
	// Test the full range of UTF-8 strings.
	err := quick.Check(roundTripsCorrectly, &quick.Config{MaxCountScale: 100.0})
	if err != nil {
		t.Error(err.Error())
	}

	// Focus on ASCII strings.
	err = quick.Check(asciiRoundTripsCorrectly, &quick.Config{MaxCountScale: 100.0})
	if err != nil {
		t.Error(err.Error())
	}
}
