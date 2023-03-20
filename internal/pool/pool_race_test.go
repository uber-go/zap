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

//go:build race

package pool_test

import (
	"sync"
	"testing"

	"go.uber.org/zap/internal/pool"
)

type pooledValue struct {
	n int64
}

func TestNew(t *testing.T) {
	// n.b. [sync.Pool] will randomly drop re-pooled objects when race is
	//      enabled, so rather than testing nondeterminsitic behavior, we use
	//      this test solely to prove that there are no races. See pool_test.go
	//      for correctness testing.

	var (
		p = pool.New(func() *pooledValue {
			return &pooledValue{
				n: -1,
			}
		})
		wg sync.WaitGroup
	)

	defer wg.Wait()

	for i := int64(0); i < 1_000; i++ {
		i := i

		wg.Add(1)
		go func() {
			defer wg.Done()

			x := p.Get()
			defer p.Put(x)

			// n.b. Must read and write the field.
			if n := x.n; n >= -1 {
				x.n = int64(i)
			}
		}()
	}
}
