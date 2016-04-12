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
	"errors"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeUser struct{ name string }

func (f fakeUser) MarshalLog(kv KeyValue) error {
	if f.name == "fail" {
		return errors.New("fail")
	}
	kv.AddString("name", f.name)
	return nil
}

func assertFieldJSON(t testing.TB, expected string, field Field) {
	enc := newJSONEncoder()
	defer enc.Free()

	field.addTo(enc)
	assert.Equal(t, expected, string(enc.bytes),
		"Unexpected JSON output after applying field %+v.", field)
}

func assertCanBeReused(t testing.TB, field Field) {
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		enc := newJSONEncoder()
		defer enc.Free()

		// Ensure using the field in multiple encoders in separate goroutines
		// does not cause any races or panics.
		wg.Add(1)
		go func() {
			defer wg.Done()
			assert.NotPanics(t, func() {
				field.addTo(enc)
			}, "Reusing a field should not cause issues")
		}()
	}

	wg.Wait()
}

func TestBoolField(t *testing.T) {
	assertFieldJSON(t, `"foo":true`, Bool("foo", true))
	assertCanBeReused(t, Bool("foo", true))
}

func TestFloat64Field(t *testing.T) {
	assertFieldJSON(t, `"foo":1.314`, Float64("foo", 1.314))
	assertCanBeReused(t, Float64("foo", 1.314))
}

func TestIntField(t *testing.T) {
	assertFieldJSON(t, `"foo":1`, Int("foo", 1))
	assertCanBeReused(t, Int("foo", 1))
}

func TestInt64Field(t *testing.T) {
	assertFieldJSON(t, `"foo":1`, Int64("foo", int64(1)))
	assertCanBeReused(t, Int64("foo", int64(1)))
}

func TestStringField(t *testing.T) {
	assertFieldJSON(t, `"foo":"bar"`, String("foo", "bar"))
	assertCanBeReused(t, String("foo", "bar"))
}

func TestStringerField(t *testing.T) {
	ip := net.ParseIP("1.2.3.4")
	assertFieldJSON(t, `"foo":"1.2.3.4"`, Stringer("foo", ip))
	assertCanBeReused(t, Stringer("foo", ip))
}

func TestTimeField(t *testing.T) {
	assertFieldJSON(t, `"foo":0`, Time("foo", time.Unix(0, 0)))
	assertCanBeReused(t, Time("foo", time.Unix(0, 0)))
}

func TestErrField(t *testing.T) {
	assertFieldJSON(t, `"error":"fail"`, Err(errors.New("fail")))
	assertCanBeReused(t, Err(errors.New("fail")))
}

func TestDurationField(t *testing.T) {
	assertFieldJSON(t, `"foo":1`, Duration("foo", time.Nanosecond))
	assertCanBeReused(t, Duration("foo", time.Nanosecond))
}

func TestObjectField(t *testing.T) {
	// Marshaling the user failed, so we expect an empty object.
	assertFieldJSON(t, `"foo":{}`, Object("foo", fakeUser{"fail"}))

	assertFieldJSON(t, `"foo":{"name":"phil"}`, Object("foo", fakeUser{"phil"}))
	assertCanBeReused(t, Object("foo", fakeUser{"phil"}))
}

func TestNestField(t *testing.T) {
	assertFieldJSON(t, `"foo":{"name":"phil","age":42}`,
		Nest("foo", String("name", "phil"), Int("age", 42)),
	)
	// Marshaling the user failed, so we expect an empty object.
	assertFieldJSON(t, `"foo":{"user":{}}`,
		Nest("foo", Object("user", fakeUser{"fail"})),
	)

	nest := Nest("foo", String("name", "phil"), Int("age", 42))
	assertCanBeReused(t, nest)
}

func TestStackField(t *testing.T) {
	enc := newJSONEncoder()
	defer enc.Free()

	Stack().addTo(enc)
	output := string(enc.bytes)

	require.True(t, strings.HasPrefix(output, `"stacktrace":`), "Stacktrace added under an unexpected key.")
	assert.Contains(t, output[13:], "zap.TestStackField", "Expected stacktrace to contain caller.")
}

func TestUnknownField(t *testing.T) {
	enc := newJSONEncoder()
	defer enc.Free()

	for _, ft := range []fieldType{unknownType, -42} {
		field := Field{fieldType: ft}
		assert.Panics(t, func() { field.addTo(enc) }, "Expected panic when using a field of unknown type.")
	}
}
