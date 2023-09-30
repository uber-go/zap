// Copyright (c) 2016-2023 Uber Technologies, Inc.
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

package zapfield

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type (
	MyKey    string
	MyValue  string
	MyValues []MyValue
)

func TestFieldConstructors(t *testing.T) {
	var (
		key    = MyKey("test key")
		value  = MyValue("test value")
		values = []MyValue{
			MyValue("test value 1"),
			MyValue("test value 2"),
		}
	)

	tests := []struct {
		name   string
		expect zap.Field
		field  zap.Field
	}{
		{"Str", zap.Field{Type: zapcore.StringType, Key: "test key", String: "test value"}, Str(key, value)},
		{"Strs", zap.Array("test key", stringArray[MyValue]{"test value 1", "test value 2"}), Strs(key, values)},
	}

	for _, tt := range tests {
		if !assert.Equal(t, tt.expect, tt.field, "Unexpected output from convenience field constructor %s.", tt.name) {
			t.Logf("type expected: %T\nGot: %T", tt.expect.Interface, tt.field.Interface)
		}
		assertCanBeReused(t, tt.field)
	}
}

func assertCanBeReused(t testing.TB, field zap.Field) {
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		enc := zapcore.NewMapObjectEncoder()

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
