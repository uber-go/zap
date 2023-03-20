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

// Package pool provides internal pool utilities.
package pool

import (
	"sync"
)

// A Constructor is a function that creates a T when a [Pool] is empty.
type Constructor[T any] func() T

// A Pool is a generic wrapper around [sync.Pool] to provide strongly-typed
// object pooling.
type Pool[T any] struct {
	pool sync.Pool
	ctor Constructor[T]
}

// New returns a new [Pool] for T, and will use ctor to construct new Ts when
// the pool is empty.
func New[T any](ctor Constructor[T]) *Pool[T] {
	return &Pool[T]{
		pool: sync.Pool{
			New: func() any {
				return ctor()
			},
		},
		ctor: ctor,
	}
}

// Get gets a new T from the pool, or creates a new one if the pool is empty.
func (p *Pool[T]) Get() T {
	if x, ok := p.pool.Get().(T); ok {
		return x
	}

	// n.b. This branch is effectively unreachable.
	return p.ctor()
}

// Put puts x into the pool.
func (p *Pool[T]) Put(x T) {
	p.pool.Put(x)
}
