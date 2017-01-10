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

// Package buffers provides strongly-typed functions to interact with a shared
// pool of byte slices.
//
// It's used heavily inside zap, but callers may also take advantage of it;
// it's particularly useful when implementing json.Marshaler, text.Marshaler,
// and similar interfaces.
package buffers

import "sync"

const _size = 1024 // create 1 KiB buffers

var _pool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 0, _size)
	},
}

// Get retrieves a slice from the pool, creating one if necessary.
// Newly-created slices have a 1 KiB capacity.
func Get() []byte {
	buf := _pool.Get().([]byte)
	return buf[:0]
}

// Put returns a slice to the pool.
func Put(buf []byte) {
	_pool.Put(buf)
}
