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
	"sync"
	"testing"
	"time"

	"go.uber.org/zap/zapcore"

	"math"

	"github.com/stretchr/testify/assert"
)

var (
	// Compiler complains about constants overflowing, so store this in a variable.
	maxUint64 = uint64(math.MaxUint64)
)

type username string

func (n username) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("username", string(n))
	return nil
}

func assertCanBeReused(t testing.TB, field zapcore.Field) {
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		enc := make(zapcore.MapObjectEncoder)

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

func TestFieldConstructors(t *testing.T) {
	// Interface types.
	fail := errors.New("fail")
	addr := net.ParseIP("1.2.3.4")
	name := username("phil")
	ints := []int{5, 6}
	nested := zapcore.Fields{String("name", "phil"), Int("age", 42)}

	tests := []struct {
		name   string
		field  zapcore.Field
		expect zapcore.Field
	}{
		{"Skip", zapcore.Field{Type: zapcore.SkipType}, Skip()},
		{"Bool", zapcore.Field{Key: "k", Type: zapcore.BoolType, Integer: 1}, Bool("k", true)},
		{"Bool", zapcore.Field{Key: "k", Type: zapcore.BoolType, Integer: 1}, Bool("k", true)},
		{"Int", zapcore.Field{Key: "k", Type: zapcore.Int64Type, Integer: 1}, Int("k", 1)},
		{"Int64", zapcore.Field{Key: "k", Type: zapcore.Int64Type, Integer: 1}, Int64("k", 1)},
		{"Uint", zapcore.Field{Key: "k", Type: zapcore.Uint64Type, Integer: 1}, Uint("k", 1)},
		{"Uint64", zapcore.Field{Key: "k", Type: zapcore.Uint64Type, Integer: 1}, Uint64("k", 1)},
		{"Uint64", zapcore.Field{Key: "k", Type: zapcore.Uint64Type, Integer: int64(maxUint64)}, Uint64("k", maxUint64)},
		{"Uintptr", zapcore.Field{Key: "k", Type: zapcore.Uint64Type, Integer: 10}, Uintptr("k", 0xa)},
		{"Error", Skip(), Error(nil)},
		{"Error", zapcore.Field{Key: "error", Type: zapcore.ErrorType, Interface: fail}, Error(fail)},
		{"String", zapcore.Field{Key: "k", Type: zapcore.StringType, String: "foo"}, String("k", "foo")},
		{"Time", zapcore.Field{Key: "k", Type: zapcore.Int64Type, Integer: 0}, Time("k", time.Unix(0, 0))},
		{"Time", zapcore.Field{Key: "k", Type: zapcore.Int64Type, Integer: 1000}, Time("k", time.Unix(1, 0))},
		{"Duration", zapcore.Field{Key: "k", Type: zapcore.Int64Type, Integer: 1}, Duration("k", time.Nanosecond)},
		{"Stringer", zapcore.Field{Key: "k", Type: zapcore.StringerType, Interface: addr}, Stringer("k", addr)},
		{"Base64", zapcore.Field{Key: "k", Type: zapcore.StringType, String: "YWIxMg=="}, Base64("k", []byte("ab12"))},
		{"Object", zapcore.Field{Key: "k", Type: zapcore.ObjectMarshalerType, Interface: name}, Object("k", name)},
		{"Reflect", zapcore.Field{Key: "k", Type: zapcore.ReflectType, Interface: ints}, Reflect("k", ints)},
		{"Nest", zapcore.Field{Key: "k", Type: zapcore.ObjectMarshalerType, Interface: nested}, Nest("k", nested...)},
		{"Any:ObjectMarshaler", Any("k", name), Object("k", name)},
		{"Any:Bool", Any("k", true), Bool("k", true)},
		{"Any:Float64", Any("k", 3.14), Float64("k", 3.14)},
		// TODO (v1.0): We could use some approximately-equal logic here, but it's
		// not worth it to test this one line. Before 1.0, we'll need to support
		// float32s explicitly, which will make this test pass.
		// {"Any:Float32", Any("k", float32(3.14)), Float32("k", 3.14)},
		{"Any:Int", Any("k", 1), Int("k", 1)},
		{"Any:Int64", Any("k", int64(1)), Int64("k", 1)},
		{"Any:Int32", Any("k", int32(1)), Int64("k", 1)},
		{"Any:Int16", Any("k", int16(1)), Int64("k", 1)},
		{"Any:Int8", Any("k", int8(1)), Int64("k", 1)},
		{"Any:Uint", Any("k", uint(1)), Uint("k", 1)},
		{"Any:Uint64", Any("k", uint64(1)), Uint64("k", 1)},
		{"Any:Uint32", Any("k", uint32(1)), Uint64("k", 1)},
		{"Any:Uint16", Any("k", uint16(1)), Uint64("k", 1)},
		{"Any:Uint8", Any("k", uint8(1)), Uint64("k", 1)},
		{"Any:Uintptr", Any("k", uintptr(1)), Uintptr("k", 1)},
		{"Any:String", Any("k", "v"), String("k", "v")},
		{"Any:Error", Any("k", errors.New("v")), String("k", "v")},
		{"Any:Time", Any("k", time.Unix(0, 0)), Time("k", time.Unix(0, 0))},
		{"Any:Duration", Any("k", time.Second), Duration("k", time.Second)},
		{"Any:Stringer", Any("k", addr), Stringer("k", addr)},
		{"Any:Fallback", Any("k", struct{}{}), Reflect("k", struct{}{})},
	}

	for _, tt := range tests {
		if !assert.Equal(t, tt.expect, tt.field, "Unexpected output from convenience field constructor %s.", tt.name) {
			t.Logf("type expected: %T\nGot: %T", tt.expect.Interface, tt.field.Interface)
		}
		assertCanBeReused(t, tt.field)
	}
}

func TestStackField(t *testing.T) {
	f := Stack()
	assert.Equal(t, "stacktrace", f.Key, "Unexpected field key.")
	assert.Equal(t, zapcore.StringType, f.Type, "Unexpected field type.")
	assert.Contains(t, f.String, "zap.TestStackField", "Expected stacktrace to contain caller.")
	assertCanBeReused(t, f)
}
