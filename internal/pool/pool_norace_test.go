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

//go:build !race

package pool_test

import (
	"bytes"
	"runtime/debug"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap/internal/pool"
)

func TestNew(t *testing.T) {
	// n.b. Disable GC to avoid the victim cache during the test.
	defer debug.SetGCPercent(debug.SetGCPercent(-1))

	p := pool.New(func() *bytes.Buffer {
		return bytes.NewBuffer([]byte("new"))
	})
	p.Put(bytes.NewBuffer([]byte(t.Name())))

	// Ensure that we always get the expected value.
	for i := 0; i < 1_000; i++ {
		func() {
			buf := p.Get()
			defer p.Put(buf)
			require.Equal(t, t.Name(), buf.String())
		}()
	}

	// Depool an extra object to ensure that the constructor is called and
	// produces an expected value.
	require.Equal(t, t.Name(), p.Get().String())
	require.Equal(t, "new", p.Get().String())
}
