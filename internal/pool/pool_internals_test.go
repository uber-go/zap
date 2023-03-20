// Copyright (c) 2023 Uber Technologies, Inc.
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

package pool

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPool_ConstructorCalledIfWrongType(t *testing.T) {
	cases := []struct {
		input      any
		expectCall bool
	}{
		{
			input:      int64(123),
			expectCall: false,
		},
		{
			input:      uint64(123),
			expectCall: true,
		},
		{
			input:      int(123),
			expectCall: true,
		},
		{
			input:      uint(123),
			expectCall: true,
		},
		{
			input:      struct{}{},
			expectCall: true,
		},
		{
			input:      nil,
			expectCall: true,
		},
	}

	for _, tt := range cases {
		t.Run(fmt.Sprintf("%T", tt.input), func(t *testing.T) {
			var (
				called bool
				pool   = New(func() int64 {
					called = true
					return 0
				})
			)

			// Override the internal pool to provide unexpected types.
			pool.pool.New = func() any {
				return tt.input
			}

			pool.pool.Put(tt.input)
			pool.Get()
			require.Equal(t, tt.expectCall, called)
		})
	}
}
