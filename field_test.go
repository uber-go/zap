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

	var out interface{}
	err := json.Unmarshal([]byte("{"+expected+"}"), &out)
	require.NoError(t, err,
		"Expected JSON snippet %q must be valid for use in an object.", expected)

	field.AddTo(enc)
	assert.Equal(t, expected, string(enc.bytes),
		"Unexpected JSON output after applying field %+v.", field)
}

func assertNotEqualFieldJSON(t testing.TB, expected string, field Field) {
	enc := newJSONEncoder()
	defer enc.Free()

	field.AddTo(enc)
	assert.NotEqual(t, expected, string(enc.bytes),
		"Unexpected JSON output after applying field %+v.", field)
}

func assertCanBeReused(t testing.TB, field Field) {
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		enc := NewJSONEncoder()
		defer enc.Free()

		// Ensure using the field in multiple encoders in separate goroutines
		// does not cause any races or panics.
		wg.Add(1)
		go func() {
			defer wg.Done()
			assert.NotPanics(t, func() {
				field.AddTo(enc)
			}, "Reusing a field should not cause issues")
		}()
	}

	wg.Wait()
}

func TestSkipField(t *testing.T) {
	assertFieldJSON(t, ``, Skip())
	assertCanBeReused(t, Skip())
}

func TestTrueBoolField(t *testing.T) {
	assertFieldJSON(t, `"foo":true`, Bool("foo", true))
	assertCanBeReused(t, Bool("foo", true))
}

func TestFalseBoolField(t *testing.T) {
	assertFieldJSON(t, `"bar":false`, Bool("bar", false))
	assertCanBeReused(t, Bool("bar", false))
}

func TestUnlikeBoolField(t *testing.T) {
	assertNotEqualFieldJSON(t, `"foo":true`, Bool("foo", false))
	assertNotEqualFieldJSON(t, `"bar":false`, Bool("bar", true))
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

func TestUintField(t *testing.T) {
	assertFieldJSON(t, `"foo":1`, Uint("foo", 1))
	assertCanBeReused(t, Uint("foo", 1))
}

func TestUint64Field(t *testing.T) {
	assertFieldJSON(t, `"foo":1`, Uint64("foo", uint64(1)))
	assertCanBeReused(t, Uint64("foo", uint64(1)))
}

func TestUintptrField(t *testing.T) {
	assertFieldJSON(t, `"foo":10`, Uintptr("foo", uintptr(0xa)))
	assertCanBeReused(t, Uintptr("foo", uintptr(0xa)))
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
	assertFieldJSON(t, `"foo":1.5`, Time("foo", time.Unix(1, int64(500*time.Millisecond))))
	assertCanBeReused(t, Time("foo", time.Unix(0, 0)))
}

func TestErrField(t *testing.T) {
	assertFieldJSON(t, `"error":"fail"`, Error(errors.New("fail")))
	assertFieldJSON(t, ``, Error(nil))
	assertCanBeReused(t, Error(errors.New("fail")))
}

func TestDurationField(t *testing.T) {
	assertFieldJSON(t, `"foo":1`, Duration("foo", time.Nanosecond))
	assertCanBeReused(t, Duration("foo", time.Nanosecond))
}

func TestMarshalerField(t *testing.T) {
	// Marshaling the user failed, so we expect an empty object and an error
	// message.
	assertFieldJSON(t, `"foo":{},"fooError":"fail"`, Marshaler("foo", fakeUser{"fail"}))

	assertFieldJSON(t, `"foo":{"name":"phil"}`, Marshaler("foo", fakeUser{"phil"}))
	assertCanBeReused(t, Marshaler("foo", fakeUser{"phil"}))
}

func TestObjectField(t *testing.T) {
	assertFieldJSON(t, `"foo":[5,6]`, Object("foo", []int{5, 6}))
	assertCanBeReused(t, Object("foo", []int{5, 6}))
}

func TestNestField(t *testing.T) {
	assertFieldJSON(t, `"foo":{"name":"phil","age":42}`,
		Nest("foo", String("name", "phil"), Int("age", 42)),
	)
	// Marshaling the user failed, so we expect an empty object and an error
	// message.
	assertFieldJSON(t, `"foo":{"user":{},"userError":"fail"}`,
		Nest("foo", Marshaler("user", fakeUser{"fail"})),
	)

	nest := Nest("foo", String("name", "phil"), Int("age", 42))
	assertCanBeReused(t, nest)
}

func TestBase64Field(t *testing.T) {
	assertFieldJSON(t, `"foo":"YWIxMg=="`,
		Base64("foo", []byte("ab12")),
	)
	assertCanBeReused(t, Base64("foo", []byte("bar")))
}

func TestLogMarshalerFunc(t *testing.T) {
	assertFieldJSON(t, `"foo":{"name":"phil"}`,
		Marshaler("foo", LogMarshalerFunc(fakeUser{"phil"}.MarshalLog)))
}

func TestStackField(t *testing.T) {
	enc := newJSONEncoder()
	defer enc.Free()

	Stack().AddTo(enc)
	output := string(enc.bytes)

	require.True(t, strings.HasPrefix(output, `"stacktrace":`), "Stacktrace added under an unexpected key.")
	assert.Contains(t, output[13:], "zap.TestStackField", "Expected stacktrace to contain caller.")
}

func TestUnknownField(t *testing.T) {
	enc := NewJSONEncoder()
	defer enc.Free()

	for _, ft := range []fieldType{unknownType, -42} {
		field := Field{fieldType: ft}
		assert.Panics(t, func() { field.AddTo(enc) }, "Expected panic when using a field of unknown type.")
	}
}
