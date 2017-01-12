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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHooks(t *testing.T) {
	tests := []struct {
		entryLevel    Level
		facilityLevel Level
		expectCall    bool
	}{
		{DebugLevel, InfoLevel, false},
		{InfoLevel, InfoLevel, true},
		{WarnLevel, InfoLevel, true},
	}

	for _, tt := range tests {
		fac, logs := NewObserver(tt.facilityLevel, 1)
		intField := makeInt64Field("foo", 42)
		ent := Entry{Message: "bar", Level: tt.entryLevel}

		var called int
		f := func(e Entry) error {
			called++
			assert.Equal(t, ent, e, "Hook called with unexpected Entry.")
			return nil
		}

		h := Hooked(fac, f)
		if ce := h.With([]Field{intField}).Check(ent, nil); ce != nil {
			ce.Write()
		}

		if tt.expectCall {
			assert.Equal(t, 1, called, "Expected to call hook once.")
			assert.Equal(
				t,
				[]ObservedLog{{Entry: ent, Context: []Field{intField}}},
				logs.AllUntimed(),
				"Unexpected logs written out.",
			)
		} else {
			assert.Equal(t, 0, called, "Didn't expect to call hook.")
			assert.Equal(t, 0, logs.Len(), "Unexpected logs written out.")
		}
	}
}
