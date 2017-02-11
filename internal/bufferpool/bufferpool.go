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

// Package bufferpool provides strongly-typed functions to interact with a shared
// pool of byte buffers.
package bufferpool

import (
	"sync"

	"go.uber.org/zap/buffer"
)

var _pool = sync.Pool{
	New: func() interface{} {
		return buffer.New()
	},
}

// Get retrieves a buffer from the pool, creating one if necessary.
func Get() *buffer.Buffer {
	buf := _pool.Get().(*buffer.Buffer)
	buf.Reset()
	return buf
}

// Put returns a slice to the pool.
func Put(buf *buffer.Buffer) {
	_pool.Put(buf)
}
