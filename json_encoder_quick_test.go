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
	"encoding/json"
	"math/rand"
	"reflect"
	"testing"
	"testing/quick"
)

func encodeString(s string) []byte {
	// Escape and quote a string using our encoder.
	enc := newJSONEncoder()
	defer enc.Free()
	enc.safeAddString(s)

	ret := make([]byte, 0, len(enc.bytes)+2)
	ret = append(ret, '"')
	ret = append(ret, enc.bytes...)
	return append(ret, '"')
}

func roundTrip(original string) bool {
	// Encode using our encoder, decode using the standard library, and assert
	// that we haven't lost any information.
	encoded := encodeString(original)

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

func roundTripASCII(s ASCII) bool {
	return roundTrip(string(s))
}

func TestJSONQuick(t *testing.T) {
	// Test the full range of UTF-8 strings.
	err := quick.Check(roundTrip, &quick.Config{MaxCountScale: 100.0})
	if err != nil {
		t.Error(err.Error())
	}

	// Focus on ASCII strings.
	err = quick.Check(roundTripASCII, &quick.Config{MaxCountScale: 100.0})
	if err != nil {
		t.Error(err.Error())
	}
}
