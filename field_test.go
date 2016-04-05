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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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

func assertPanicsIfReused(t testing.TB, field Field) {
	enc := newJSONEncoder()
	defer enc.Free()

	field.addTo(enc)
	assert.Panics(t, func() { field.addTo(enc) }, "Expected reusing field to panic.")
}

func assertCanBeReused(t testing.TB, field Field) {
	enc := newJSONEncoder()
	defer enc.Free()

	field.addTo(enc)
	assert.NotPanics(t, func() { field.addTo(enc) }, "Expected reusing field to panic.")
}

func TestBoolField(t *testing.T) {
	assertFieldJSON(t, `"foo":true`, Bool("foo", true))
	assertPanicsIfReused(t, Bool("foo", true))

	assertFieldJSON(t, `"foo":true`, Bool("foo", true, Keep))
	assertCanBeReused(t, Bool("foo", true, Keep))
}

func TestFloat64Field(t *testing.T) {
	assertFieldJSON(t, `"foo":1.314`, Float64("foo", 1.314))
	assertPanicsIfReused(t, Float64("foo", 1.314))

	assertFieldJSON(t, `"foo":1.314`, Float64("foo", 1.314, Keep))
	assertCanBeReused(t, Float64("foo", 1.314, Keep))
}

func TestIntField(t *testing.T) {
	assertFieldJSON(t, `"foo":1`, Int("foo", 1))
	assertPanicsIfReused(t, Int("foo", 1))

	assertFieldJSON(t, `"foo":1`, Int("foo", 1, Keep))
	assertCanBeReused(t, Int("foo", 1, Keep))
}

func TestInt64Field(t *testing.T) {
	assertFieldJSON(t, `"foo":1`, Int64("foo", int64(1)))
	assertPanicsIfReused(t, Int64("foo", int64(1)))

	assertFieldJSON(t, `"foo":1`, Int64("foo", int64(1), Keep))
	assertCanBeReused(t, Int64("foo", int64(1), Keep))
}

func TestStringField(t *testing.T) {
	assertFieldJSON(t, `"foo":"bar"`, String("foo", "bar"))
	assertPanicsIfReused(t, String("foo", "bar"))

	assertFieldJSON(t, `"foo":"bar"`, String("foo", "bar", Keep))
	assertCanBeReused(t, String("foo", "bar", Keep))
}

func TestTimeField(t *testing.T) {
	assertFieldJSON(t, `"foo":0`, Time("foo", time.Unix(0, 0)))
	assertPanicsIfReused(t, Time("foo", time.Unix(0, 0)))

	assertFieldJSON(t, `"foo":0`, Time("foo", time.Unix(0, 0), Keep))
	assertCanBeReused(t, Time("foo", time.Unix(0, 0), Keep))
}

func TestErrField(t *testing.T) {
	assertFieldJSON(t, `"error":"fail"`, Err(errors.New("fail")))
	assertPanicsIfReused(t, Err(errors.New("fail")))

	assertFieldJSON(t, `"error":"fail"`, Err(errors.New("fail"), Keep))
	assertCanBeReused(t, Err(errors.New("fail"), Keep))
}

func TestDurationField(t *testing.T) {
	assertFieldJSON(t, `"foo":1`, Duration("foo", time.Nanosecond))
	assertPanicsIfReused(t, Duration("foo", time.Nanosecond))

	assertFieldJSON(t, `"foo":1`, Duration("foo", time.Nanosecond, Keep))
	assertCanBeReused(t, Duration("foo", time.Nanosecond, Keep))
}

func TestObjectField(t *testing.T) {
	assertFieldJSON(t, `"foo":{"name":"phil"}`, Object("foo", fakeUser{"phil"}))
	// Marshaling the user failed, so we expect an empty object.
	assertFieldJSON(t, `"foo":{}`, Object("foo", fakeUser{"fail"}))
	assertPanicsIfReused(t, Object("foo", fakeUser{"phil"}))

	assertFieldJSON(t, `"foo":{"name":"phil"}`, Object("foo", fakeUser{"phil"}, Keep))
	assertCanBeReused(t, Object("foo", fakeUser{"phil"}, Keep))
}

func TestNestField(t *testing.T) {
	assertFieldJSON(t, `"foo":{"name":"phil","age":42}`,
		Nest("foo", String("name", "phil"), Int("age", 42)),
	)
	// Marshaling the user failed, so we expect an empty object.
	assertFieldJSON(t, `"foo":{"user":{}}`,
		Nest("foo", Object("user", fakeUser{"fail"})),
	)
	assertPanicsIfReused(t, Nest("foo", String("name", "phil"), Int("age", 42)))

	nest := Nest("foo", String("name", "phil"), Int("age", 42))
	Keep(nest)
	assertCanBeReused(t, nest)

}
